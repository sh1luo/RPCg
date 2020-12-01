package loadbalance

import (
	"errors"
	"fmt"
	"sync"
)

type Manager struct {
	l sync.Mutex
	balancer map[string]Balancer
}

func (m *Manager) registerBalancer(name string, balancer Balancer) error {
	if _,ok := m.balancer[name];ok {
		return errors.New("this balancer already exist")
	}
	m.balancer[name] = balancer
	return nil
}

type Balancer interface {
	Add(node ...KV) error
	Remove(node KV) error
	Get(key KV) (string, bool)
}

// KV is to compatible with complex types, as long as implementing the KV interface,
// which is consistent with fmt.Stringer
type KV interface {
	String() string
}

type Instance struct {
	Host          string
	Port          uint16
	Weight        uint32
	CurrentWeight uint32
}

func (i Instance) String() string {
	return fmt.Sprintf("%s:%d", i.Host, i.Port)
}