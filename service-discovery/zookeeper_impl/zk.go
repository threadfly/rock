package zookeeper_impl

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"rock/log"
	sd "rock/service-discovery"

	"github.com/go-zookeeper/zk"
)

var (
	ErrNodeEmpty = errors.New("node is empty")
)

type ThirdAction interface {
	Decode([]byte) (node interface{}, err error)                             // decode from zk node content
	Wrap(si sd.SIndex, nodes []interface{}) (endpoint sd.SEndpoint, ok bool) //
}

type customSIndex string

func (cs customSIndex) Key() string {
	return string(cs)
}

func tryTimes(times int, f func() error) (err error) {
	for i := 0; i < times; i++ {
		err = f()
		if err != nil {
			log.Errorf("tryTimes, No.%d/%d, %s", i, times, err)
			time.Sleep(time.Duration((i+1)*200) * time.Millisecond)
			continue
		}
		return
	}
	return
}

type zooKeeperConnManager struct {
	addrs []string
	rec   bool
	c     *zk.Conn
	e     <-chan zk.Event
}

func (zcm *zooKeeperConnManager) conn() (*zk.Conn, error) {
	if zcm.rec {
		c, e, err := zk.Connect(zcm.addrs, 10*time.Second)
		if err != nil {
			return nil, err
		}
		if zcm.c != nil {
			gcc := zcm.c
			go func() {
				time.Sleep(time.Second * 30)
				gcc.Close()
			}()
		}
		zcm.c = c
		zcm.e = e
		zcm.rec = false
		return c, err
	} else {
		return zcm.c, nil
	}
}

func (zcm *zooKeeperConnManager) reconn() (*zk.Conn, error) {
	zcm.rec = true
	return zcm.conn()
}

func (zcm *zooKeeperConnManager) reload(addrs []string) {
	zcm.addrs = addrs
	zcm.rec = true
}

type zooKeeperNameSpace struct {
	key2path map[string]string
}

func (zns *zooKeeperNameSpace) reload(key2path map[string]string) {
	zns.key2path = key2path
}

func (zns *zooKeeperNameSpace) query(key string) (path string, exist bool) {
	p, ok := zns.key2path[key]
	return p, ok
}

func (zns *zooKeeperNameSpace) foreach(errAbort bool, f func(k, v string) error) (err error) {
	for key, path := range zns.key2path {
		tryFunc := func() error {
			return f(key, path)
		}
		err = tryTimes(1, tryFunc)
		if err != nil {
			log.Errorf("zookeeper namespace foreach %s => %s, %s", key, path, err)
			if errAbort {
				return
			}
		}
	}
	return
}

type zooKeeperWatch struct {
	zcm   *zooKeeperConnManager
	zns   *zooKeeperNameSpace
	watch chan sd.Event
}

func NewZooKeeperWatch(zcm *zooKeeperConnManager, zns *zooKeeperNameSpace) *zooKeeperWatch {
	return &zooKeeperWatch{
		zcm:   zcm,
		zns:   zns,
		watch: make(chan sd.Event, 8),
	}
}

func (zkw *zooKeeperWatch) init() {
	go zkw.watchd()
}

func (zkw *zooKeeperWatch) watchd() {
	var wg sync.WaitGroup
	for {
		if zkw.zcm.e == nil {
			log.Infof("zk event have no init")
			time.Sleep(time.Second)
			continue
		}

		connEvent := zkw.zcm.e
		select {
		case e, ok := <-connEvent:
			if !ok {
				log.Infof("conn event have closed")
				continue
			}
			switch e.State {
			case zk.StateDisconnected, zk.StateUnknown:
				continue
			}
		default:
		}

		wg.Add(len(zkw.zns.key2path))
		for k, v := range zkw.zns.key2path {
			go func(key, path string) {
			xyz:
				for {
					_, _, event, err := zkw.zcm.c.ChildrenW(path)
					if err != nil {
						log.Errorf("children %s %s", path, err)
					}

					switch err {
					case zk.ErrConnectionClosed, zk.ErrClosing, zk.ErrSessionMoved, zk.ErrUnknown, zk.ErrNoServer:
						log.Errorf("exit children %s %s", path, err)
						time.Sleep(time.Second)
						break xyz
					default:
					}

					e, ok := <-event
					if !ok {
						break
					}
					switch e.Type {
					//case zk.EventNodeDeleted:
					//	log.Infof("watch %s:%s deleted", key, path)
					//	zkw.watch <- sd.Event{
					//		Sindex: customSIndex(key),
					//		Typ:    sd.EventRemove,
					//	}
					//case zk.EventNodeCreated:
					//	log.Infof("watch %s:%s create", key, path)
					//	zkw.watch <- sd.Event{
					//		Sindex: customSIndex(key),
					//		Typ:    sd.EventAdd,
					//	}
					case zk.EventNodeChildrenChanged:
						log.Infof("watch %s:%s child changed", key, path)
						zkw.watch <- sd.Event{
							Sindex: customSIndex(key),
							Typ:    sd.EventChildChange,
						}
					}
				}
				wg.Done()
				time.Sleep(200 * time.Millisecond)
			}(k, v)
		}
		wg.Wait()
	}
}

type ZooKeeperSource struct {
	zcm      *zooKeeperConnManager
	zns      *zooKeeperNameSpace
	zkw      *zooKeeperWatch
	ta       ThirdAction
	lkupPool *sync.Pool
}

func NewZooKeeperSource(key2path map[string]string, addrs []string, ta ThirdAction) *ZooKeeperSource {
	result := &ZooKeeperSource{
		zcm: &zooKeeperConnManager{},
		zns: &zooKeeperNameSpace{},
		ta:  ta,
	}

	result.zkw = NewZooKeeperWatch(result.zcm, result.zns)

	result.zcm.reload(addrs)
	result.zns.reload(key2path)
	result.lkupPool = &sync.Pool{
		New: func() interface{} {
			return &lookupContext{
				zks:   result,
				nodes: make([]interface{}, 0, 3),
			}
		},
	}
	return result
}

func (zks *ZooKeeperSource) Reload(key2path map[string]string, addrs []string) {
	zks.zns.reload(key2path)
	zks.zcm.reload(addrs)
}

func (zks *ZooKeeperSource) Init() error {
	_, err := zks.zcm.conn()
	if err != nil {
		return err
	}

	zks.zkw.init()
	return err
}

func (zks *ZooKeeperSource) getAvailConn() (*zk.Conn, error) {
	c, err := zks.zcm.conn()
	if err != nil {
		log.Errorf("get zookeeper conn, %s", err)
		c, err = zks.zcm.reconn()
		if err != nil {
			log.Errorf("reconnect zookeeper, %s", err)
			return nil, err
		}
	}
	return c, nil
}

func (zks *ZooKeeperSource) Get(si sd.SIndex) (sd.SEndpoint, bool) {
	path, exist := zks.zns.query(si.Key())
	if !exist {
		log.Infof("get key:%s 's path not exist", si.Key())
		return nil, exist
	}
	se, err := zks.get(si.Key(), path)
	if err != nil {
		log.Errorf("Get zookeeper si:%s, %s", si.Key(), err)
		switch err {
		case zk.ErrConnectionClosed, zk.ErrClosing,
			zk.ErrSessionMoved, zk.ErrUnknown:
			return nil, true
		default:
			return nil, false
		}
	}
	return se, true
}

func (zks *ZooKeeperSource) get(key, path string) (sd.SEndpoint, error) {
	c, err := zks.getAvailConn()
	if err != nil {
		log.Errorf("get avail conn, %s", err)
		return nil, err
	}

	ctx := zks.lkupPool.Get().(*lookupContext)
	ctx.c = c
	defer func() {
		ctx.c = nil
		ctx.nodes = ctx.nodes[:0]
		zks.lkupPool.Put(ctx)
	}()

	err = ctx.walkPath(path)
	if err != nil {
		log.Errorf("get zookeeper path:%s, %s", path, err)
		if len(ctx.nodes) <= 0 {
			return nil, err
		}
		log.Infof("get zookeeper path:%s, found %d nodes, so continue", path, len(ctx.nodes))
	} else if len(ctx.nodes) == 0 {
		return nil, ErrNodeEmpty
	}

	se, ok := ctx.zks.ta.Wrap(customSIndex(key), ctx.nodes)
	if ok {
		return se, nil
	} else {
		return nil, fmt.Errorf("third action wrap not succ")
	}
}

func (zks *ZooKeeperSource) FetchAll() ([]sd.SEndpoint, error) {
	var result []sd.SEndpoint
	zks.zns.foreach(false, func(key, path string) error {
		se, err := zks.get(key, path)
		if err != nil {
			log.Infof("zookeeper service source fetchall, key:%s => path:%s, %s", key, path, err)
			return err
		}

		if result == nil {
			result = make([]sd.SEndpoint, 0, 3)
		}
		result = append(result, se)
		return nil
	})

	return result, nil
}

func (zks *ZooKeeperSource) Watch() <-chan sd.Event {
	return zks.zkw.watch
}

type lookupContext struct {
	zks   *ZooKeeperSource
	c     *zk.Conn
	nodes []interface{}
}

func (ctx *lookupContext) checkTryReconn(err error) error {
	switch err {
	case zk.ErrConnectionClosed, zk.ErrClosing, zk.ErrSessionMoved:
		ctx.c, err = ctx.zks.zcm.reconn()
		return err
	default:
		return err
	}
}

func (ctx *lookupContext) readdir(path string) ([]string, error) {
	cs, _, err := ctx.c.Children(path)
	return cs, ctx.checkTryReconn(err)
}

func (ctx *lookupContext) lookup(path string) error {
	bytes, _, err := ctx.c.Get(path)
	if err != nil {
		return ctx.checkTryReconn(err)
	}

	node, err := ctx.zks.ta.Decode(bytes)
	if err != nil {
		log.Errorf("path:%s decode bytes:%s", path, string(bytes))
		return err
	}

	if ctx.nodes == nil {
		ctx.nodes = make([]interface{}, 0, 3)
	} else {
		ctx.nodes = append(ctx.nodes, node)
	}
	return nil
}

func (ctx *lookupContext) walkPath(path string) (err error) {
	var childs []string
	readdir := func() error {
		x, err := ctx.readdir(path)
		childs = x
		return err
	}

	err = tryTimes(3, readdir)
	if err != nil {
		return err
	}

	if len(childs) == 0 {
		lookup := func() error {
			return ctx.lookup(path)
		}
		return tryTimes(3, lookup)
	} else {
		for _, base := range childs {
			err = ctx.walkPath(path + "/" + base)
			if err != nil {
				log.Errorf("walkPath %s, %s", path+"/"+base, err)
			}
		}
	}
	return
}
