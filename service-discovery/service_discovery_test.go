package service_discovery

import (
	"log"
	"strconv"
	"testing"
	"time"
)

type SI int8

var (
	sixmap map[int8]string = make(map[int8]string)
)

func (si SI) Key() string {
	return sixmap[int8(si)]
}

type SE struct {
	si SIndex
	v  int
}

func (se *SE) Key() SIndex {
	return se.si
}

func (se *SE) Value() interface{} {
	return se.v
}

func (se *SE) Eq(se2 SEndpoint) bool {
	x, _ := se2.Value().(int)
	return se.v == x
}

type MockSSource struct {
	cache     map[string]*SE
	missBegin time.Time
	missEnd   time.Time
	e         chan Event
}

func (ms *MockSSource) Init() error {
	ms.cache = make(map[string]*SE)
	for i := 0; i <= 127; i++ {
		sixmap[int8(i)] = strconv.FormatUint(uint64(i), 10)
		if i%2 == 0 {
			ms.cache[SI(i).Key()] = &SE{
				si: SI(i),
				v:  i,
			}
		}
	}
	ms.missBegin = time.Now().Add(30 * time.Second)
	ms.missEnd = time.Now().Add(60 * time.Second)
	ms.e = make(chan Event)
	return nil
}

func (ms *MockSSource) Get(si SIndex) (SEndpoint, bool) {
	log.Printf("mock source get si:%s\n", si.Key())
	now := time.Now()
	if now.After(ms.missBegin) && now.Before(ms.missEnd) {
		return nil, false
	} else {
		return &SE{
			si: si.(SI),
			v:  int(si.(SI)),
		}, true
	}
}

func (ms *MockSSource) FetchAll() ([]SEndpoint, error) {
	result := make([]SEndpoint, 0, len(ms.cache))
	for _, se := range ms.cache {
		result = append(result, se)
	}
	return result, nil
}

func (ms *MockSSource) Watch() <-chan Event {
	return ms.e
}

func TestServiceDiscovery(t *testing.T) {
	ms := &MockSSource{}
	err := ms.Init()
	if err != nil {
		t.Errorf("init , %s", err)
	}

	sd, err := NewServiceDiscovery(ms, time.Second*15)
	if err != nil {
		t.Errorf("new sd, %s", err)
	}

	var step int

	f := func() {
		for {
			step = 1
			for i := 0; i <= 127; i += step {
				se, _ := sd.Get(SI(i))
				if se == nil {
					log.Printf("si:%d not found", i)
				}
			}

			step = 2
			for i := 0; i <= 127; i += step {
				se, _ := sd.Get(SI(i))
				if se == nil {
					log.Printf("si:%d not found", i)
				}
			}

			step = 3
			for i := 0; i <= 127; i += step {
				se, _ := sd.Get(SI(i))
				if se == nil {
					log.Printf("si:%d not found", i)
				}
			}

			step = 4
			for i := 0; i <= 127; i += step {
				se, _ := sd.Get(SI(i))
				if se == nil {
					log.Printf("si:%d not found", i)
				}
			}
		}
	}
	go f()
	f()
}
