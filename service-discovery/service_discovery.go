package service_discovery

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"rock/log"
)

var (
	ErrNoFound = errors.New("no found service endpoint")
	ErrAssert  = errors.New("assert type")
	ErrUnknown = errors.New("unknown")
)

// service index
type SIndex interface {
	Key() string
}

// service endpoint
type SEndpoint interface {
	Key() SIndex
	Value() interface{}
	Eq(SEndpoint) bool
}

type EventType uint8

const (
	EventRemove EventType = iota
	EventAdd
	EventChildChange
)

type Event struct {
	Sindex SIndex
	Typ    EventType
}

type SSource interface {
	Init() error
	// serial request
	// When bool is true and x is nil, it proves that ssource is faulty,
	// causing the query to fail, so there is no need to update
	Get(SIndex) (SEndpoint, bool)
	FetchAll() ([]SEndpoint, error)
	Watch() <-chan Event
}

type valueTyp uint8

const (
	miss valueTyp = iota
	comm
	get
)

/*
																		watch
		---------------------------------------------------------------------------------------------------------
		|																														   			|									|									|
		|					 probe								     get                            V									V									V
	donothing ---------------> mark ---------------->  update			   		 add						  remove						change
		^																									|
		|                   												      |
		---------------------------------------------------
																					update
*/
type stateTyp int32

const (
	donothing stateTyp = iota
	mark
	update
	add
	remove
	change
)

type value struct {
	si  SIndex
	se  SEndpoint
	typ valueTyp
	ttl time.Time
	sync.Mutex
	update int32 // 0:do nothing, 1:update 2:mark
}

type SDiscovery struct {
	ss  SSource   // 获取服务点的源
	ct  *sync.Map // for fast path
	ttl time.Duration
	c   chan *value
}

func NewServiceDiscovery(ss SSource, ttl time.Duration) (*SDiscovery, error) {
	err := ss.Init()
	if err != nil {
		return nil, err
	}

	if ttl < 10*time.Second || ttl > 30*time.Second {
		ttl = 30 * time.Second
	}

	sd := &SDiscovery{
		ss:  ss,
		ct:  new(sync.Map),
		ttl: ttl,
		c:   make(chan *value, 512),
	}

	return sd, sd.init()
}

func (sd *SDiscovery) init() error {
	se, err := sd.ss.FetchAll()
	if err != nil {
		return err
	}

	for _, e := range se {
		sd.ct.Store(e.Key().Key(), &value{
			si:  e.Key(),
			se:  e,
			typ: comm,
			ttl: time.Now().Add(sd.ttl),
		})
	}
	go sd.probe()
	go sd.update()
	go sd.watch()
	return nil
}

func (sd *SDiscovery) watch() {
	for e := range sd.ss.Watch() {
		val := sd.get(e.Sindex)

		var succ bool
		switch e.Typ {
		case EventRemove:
			if val == nil {
				continue
			}
			succ = atomic.CompareAndSwapInt32(&(val.update), int32(donothing), int32(remove))
		case EventAdd:
			if val == nil {
				val = &value{
					si: e.Sindex,
				}
			}
			succ = atomic.CompareAndSwapInt32(&(val.update), int32(donothing), int32(add))
		case EventChildChange:
			if val == nil {
				val = &value{
					si: e.Sindex,
				}
			}
			succ = atomic.CompareAndSwapInt32(&(val.update), int32(donothing), int32(change))
		}

		if succ {
			log.Infof("watch key:%s, %d type", e.Sindex.Key(), e.Typ)
			sd.c <- val
		}
	}
}

func (sd *SDiscovery) get(si SIndex) *value {
	v, ok := sd.ct.Load(si.Key())
	if !ok {
		return nil
	}

	val, ok := v.(*value)
	if !ok {
		log.Errorf("assert type error")
		return nil
	}
	return val
}

func (sd *SDiscovery) Get(si SIndex) (SEndpoint, error) {
	f := func(vx *value) (se SEndpoint, err error, expired bool) {
		expired = vx.update == int32(mark)
		se = vx.se
		if se == nil {
			return se, ErrNoFound, expired
		}

		switch vx.typ {
		case miss:
			return nil, ErrNoFound, expired
		case comm:
			return se, nil, expired
		}
		return nil, ErrUnknown, false
	}

Load:
	val := sd.get(si)
	if val != nil {
		// fast path
		se, err, exp := f(val)
		if exp {
			succ := atomic.CompareAndSwapInt32(&(val.update), int32(mark), int32(update))
			if succ {
				sd.c <- val
			}
		}
		if err != nil {
			return nil, err
		}
		return se, nil
	}

	// slow path
	val = &value{
		si:  si,
		typ: get,
	}

	var err error
	val.Lock()
	v, load := sd.ct.LoadOrStore(si.Key(), val)
	target := v.(*value)
	if load {
		// other writed
		val.Unlock()
		target.Lock()
		target.Unlock()
		goto Load
	} else {
		val.ttl = time.Now().Add(sd.ttl)
		se, ok := sd.ss.Get(si)
		if !ok || (ok && se == nil) {
			val.typ = miss
			err = ErrNoFound
		} else {
			val.se = se
			val.typ = comm
		}
	}

	val.Unlock()
	return val.se, err
}

// update container
func (sd *SDiscovery) probe() {
	ticker := time.NewTicker(time.Second * 2)
	for {
		t := <-ticker.C
		sd.ct.Range(func(key, val interface{}) bool {
			vx := val.(*value)
			if vx.ttl.After(t) {
				// 没过期跳过
				return true
			}
			switch vx.typ {
			case miss, comm:
				// filter get
				atomic.CompareAndSwapInt32(&(vx.update), int32(donothing), int32(mark))
			}
			return true
		})
	}
}

func (sd *SDiscovery) disturb() time.Duration {
	multi := rand.Intn(10)
	if multi == 0 {
		multi = 1
	}
	return sd.ttl * time.Duration(multi) / time.Duration(10)
}

func (sd *SDiscovery) update() {
	for {
		v := <-sd.c
		if v == nil {
			log.Errorf("v is nil")
			continue
		}

		switch stateTyp(v.update) {
		case update, remove, add, change:
			se, ok := sd.ss.Get(v.si)
			if ok && se != nil {
				v.typ = comm
				v.se = se
				v.ttl = time.Now().Add(sd.ttl + sd.disturb()) //disturb
			} else {
				v.typ = miss
				v.se = nil
				v.ttl = time.Now().Add((sd.ttl / time.Duration(2)) + sd.disturb()) // for fast check
			}
		}

		atomic.CompareAndSwapInt32(&(v.update), v.update, int32(donothing))
		// Prevent excessive concurrency, although it is a bit frustrating, but simple and effective
		time.Sleep(100 * time.Millisecond)
	}
}
