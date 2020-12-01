package consistenthash

import (
	"errors"
	"fmt"
	"hash/crc32"
	"rpcg/loadbalance"
	"sort"
	"sync"
)

type HashFunc func(data []byte) uint32

type ChMap struct {
	l            *sync.RWMutex

	// use dependency injection mode
	hashFunc     HashFunc

	replicates   int
	sortedKeys   []uint32
	hashRing     map[uint32]loadbalance.KV
	NodeExistMap map[loadbalance.KV]bool
}

var _ loadbalance.Balancer = (*ChMap)(nil)

// NewChMap create a ChMap which has replicates of virtual nodes and custom hash function.
func NewChMap(replicates int, fn HashFunc) *ChMap {
	c := &ChMap{
		l:            &sync.RWMutex{},
		hashFunc:     fn,
		replicates:   replicates,
		hashRing:     make(map[uint32]loadbalance.KV),
		NodeExistMap: make(map[loadbalance.KV]bool),
	}
	if c.hashFunc == nil {
		c.hashFunc = crc32.ChecksumIEEE // by default crc32 is used
	}
	return c
}

// Add add the node to the hash ring
func (c *ChMap) Add(nodes ...loadbalance.KV) error {
	c.l.Lock()
	defer c.l.RUnlock()
	for _, node := range nodes {
		if _, hit := c.NodeExistMap[node]; hit {
			return errors.New("node already exist")
		}
		for i := 0; i < c.replicates; i++ {
			virtualHost := fmt.Sprintf("%s#%d", node.String(), i)
			virtualKey := c.hashFunc([]byte(virtualHost))
			c.hashRing[virtualKey] = node
			c.sortedKeys = append(c.sortedKeys, virtualKey)
		}
		c.NodeExistMap[node] = true
		sort.Slice(c.sortedKeys, func(i, j int) bool {
			return c.sortedKeys[i] < c.sortedKeys[j]
		})
	}
	return nil
}

// Remove remove the node and all the virtual nodes from the key
func (c *ChMap) Remove(node loadbalance.KV) error {
	panic("implement me")
}

func (c *ChMap) Get(key loadbalance.KV) (string, bool) {
	panic("implement me")
}

func (c *ChMap) rebuildHashRing() {

}
