package zookeeper_impl

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"rock/log"
	sd "rock/service-discovery"

	"github.com/go-zookeeper/zk"
)

const (
	MGGW  = "mongogw"
	MGGW1 = "mongogw1"
	MGGW2 = "mongogw2"
	MGGW3 = "mongogw3"
)

type Node struct {
	Ip   string
	Port int
}

// implementing of SEndpoint
type Endpoint struct {
	si    sd.SIndex
	nodes []*Node
	i     int
}

func (e *Endpoint) Key() sd.SIndex {
	return e.si
}

func (e *Endpoint) Value() interface{} {
	return e
}

func (e *Endpoint) Eq(se sd.SEndpoint) bool {
	v, ok := se.(*Endpoint)
	if !ok {
		return false
	}
	return e.Key() == v.Key()
}

func (e *Endpoint) Get() (ip string, port int, err error) {
	if len(e.nodes) == 0 {
		return ip, port, fmt.Errorf("not found ip&port")
	}

	if e.i >= len(e.nodes) {
		e.i = 0
	}

	ip = e.nodes[e.i].Ip
	port = e.nodes[e.i].Port
	e.i++
	return
}

type Json3rdAction struct {
}

func (a *Json3rdAction) Decode(bytes []byte) (node interface{}, err error) {
	node = new(Node)
	return node, json.Unmarshal(bytes, node)
}

func (a *Json3rdAction) Wrap(si sd.SIndex, nodes []interface{}) (endpoint sd.SEndpoint, ok bool) {
	ep := &Endpoint{
		si:    si,
		nodes: make([]*Node, 0, 3),
	}

	for i, v := range nodes {
		n, ok := v.(*Node)
		if !ok {
			fmt.Printf("i:%d assert faild, node:%#v", i, v)
		} else {
			ep.nodes = append(ep.nodes, n)
		}
	}
	return ep, true
}

type SimulationServiceRegister struct {
	c *zk.Conn
}

func (ssr *SimulationServiceRegister) Init(addrs []string) error {
	c, e, err := zk.Connect(addrs, time.Second*10)
	if err != nil {
		return err
	}

	go func() {
		for {
			event, ok := <-e
			if !ok {
				return
			}
			log.Infof("event, %s", event)
		}
	}()

	ssr.c = c
	return nil
}

func (ssr *SimulationServiceRegister) Register(path string, v interface{}) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	acl := zk.WorldACL(zk.PermAll)
	log.Infof("before register path:%s ,bytes:%s", path, string(bytes))
	path, err = ssr.c.Create(path, bytes, zk.FlagEphemeral, acl)
	if err != nil {
		return err
	}
	log.Infof("after register path:%s ", path)

	return err
}

func TestZookeeperImpl(t *testing.T) {
	key2path := map[string]string{
		MGGW:  "/NS/x/y",
		MGGW1: "/NS/x/y",
		MGGW2: "/NS/x/y",
		MGGW3: "/NS/x/y",
	}
	addrs := []string{"ipv41:2181", "ipv42:2181", "ipv43:2181"}
	ssr := &SimulationServiceRegister{}
	err := ssr.Init(addrs)
	if err != nil {
		t.Errorf("simulation init, %s", err)
		return
	}

	n0 := &Node{
		Ip:   "199.199.199.199",
		Port: 1999,
	}
	err = ssr.Register(key2path[MGGW]+"/n0", n0)
	if err != nil {
		t.Errorf("register n0, %s", err)
		return
	}

	n1 := &Node{
		Ip:   "200.200.200.200",
		Port: 2000,
	}
	err = ssr.Register(key2path[MGGW]+"/n1", n1)
	if err != nil {
		t.Errorf("register n1, %s", err)
		return
	}

	err = ssr.Register(key2path[MGGW1]+"/n11", n1)
	if err != nil {
		t.Errorf("register n11, %s", err)
		return
	}

	err = ssr.Register(key2path[MGGW2]+"/n22", n1)
	if err != nil {
		t.Errorf("register n22, %s", err)
		return
	}

	err = ssr.Register(key2path[MGGW3]+"/n33", n1)
	if err != nil {
		t.Errorf("register n33, %s", err)
		return
	}

	ac := &Json3rdAction{}
	zks := NewZooKeeperSource(key2path, addrs, ac)
	zksd, err := sd.NewServiceDiscovery(zks, 300*time.Second)
	if err != nil {
		t.Errorf("new service discovery, %s", err)
		return
	}

	var (
		tryTimes int
		se       sd.SEndpoint
		err1     error
	)

	for {
		se, err1 = zksd.Get(customSIndex(MGGW))
		if err1 != nil {
			t.Errorf("zksd get %s , %s", MGGW, err1)
			tryTimes++
			if tryTimes >= 3 {
				return
			}
			continue
		}
		t.Logf("zksd get %s, se is %#v, err is %#v", MGGW, se, err1)
		break
	}

	rv := se.Value().(*Endpoint)
	ip, port, err := rv.Get()
	if err != nil {
		t.Errorf("1 Get, %s", err)
	} else {
		t.Logf("1 Get ip:%s port:%d , succ", ip, port)
	}
	ip, port, err = rv.Get()
	if err != nil {
		t.Errorf("2 Get, %s", err)
	} else {
		t.Logf("2 Get ip:%s port:%d , succ", ip, port)
	}
	ip, port, err = rv.Get()
	if err != nil {
		t.Errorf("3 Get, %s", err)
	} else {
		t.Logf("3 Get ip:%s port:%d , succ", ip, port)
	}
	ip, port, err = rv.Get()
	if err != nil {
		t.Errorf("4 Get, %s", err)
	} else {
		t.Logf("4 Get ip:%s port:%d , succ", ip, port)
	}

	ip, port, err = rv.Get()
	if err != nil {
		t.Errorf("5 Get, %s", err)
	} else {
		t.Logf("5 Get ip:%s port:%d , succ", ip, port)
	}

	ip, port, err = rv.Get()
	if err != nil {
		t.Errorf("6 Get, %s", err)
	} else {
		t.Logf("6 Get ip:%s port:%d , succ", ip, port)
	}

	for {
		se, err1 = zksd.Get(customSIndex(MGGW2))
		if err1 != nil {
			t.Errorf("zksd get %s , %s", MGGW, err1)
			tryTimes++
			if tryTimes >= 3 {
				return
			}
			continue
		}
		t.Logf("zksd get %s, se is %#v, succ", MGGW, se)
		tryTimes++
		if tryTimes >= 1000000000 {
			break
		}

		rv = se.Value().(*Endpoint)
		ip, port, err = rv.Get()
		if err != nil {
			t.Errorf("Get, %s", err)
		} else {
			t.Logf("Get ip:%s port:%d , succ", ip, port)
		}
		time.Sleep(time.Second)
	}

	se, err1 = zksd.Get(customSIndex(MGGW3))
	if err1 != nil {
		log.Errorf("last get , %s", err1)
		return
	}
	rv = se.Value().(*Endpoint)
	ip, port, err = rv.Get()
	if err != nil {
		t.Errorf("7 Get, %s", err)
	} else {
		t.Logf("7 Get ip:%s port:%d , succ", ip, port)
	}
}
