package xclient

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"rpcg/utils"
	"sort"
	"sync"
)

type HashFunc func(data []byte) uint32

type ChMap struct {
	l *sync.RWMutex

	// use dependency injection mode
	hashFunc HashFunc

	replicates   int
	sortedKeys   []uint32
	hashRing     map[uint32]string
	NodeExistMap map[string]bool
}

var _ Balancer = (*ChMap)(nil)

// newConsistentHashBalancer create a ChMap which has replicates of virtual nodes and custom hash function.
// TODO:using params to customize newConsistentHashBalancer
func newConsistentHashBalancer() *ChMap {
	c := &ChMap{
		l:            &sync.RWMutex{},
		replicates:   128, // to be optimized
		hashRing:     make(map[uint32]string),
		NodeExistMap: make(map[string]bool),
	}
	if c.hashFunc == nil {
		c.hashFunc = crc32.ChecksumIEEE // by default crc32 is used
	}
	return c
}

// TODO:The selection method needs to be optimized
func (c *ChMap) Pick() string {
	c.l.RLock()
	defer c.l.RUnlock()
	n := utils.Intn(2 ^ 32 - 1)
	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(n))
	hashKey := c.hashFunc(key)

	i := sort.Search(len(c.sortedKeys), func(i int) bool {
		return c.sortedKeys[i] >= hashKey
	})

	return c.hashRing[c.sortedKeys[i%len(c.sortedKeys)]]
}

func (c *ChMap) UpdateAllServers(servers map[string]string) {
	c.l.Lock()
	defer c.l.Unlock()
	for s := range servers {
		if _, hit := c.NodeExistMap[s]; hit {
			return
		}
		for i := 0; i < c.replicates; i++ {
			virtualNode := fmt.Sprintf("%s##%v", s, i)
			virtualKey := c.hashFunc([]byte(virtualNode))
			c.hashRing[virtualKey] = s
			c.sortedKeys = append(c.sortedKeys, virtualKey)
		}
		c.NodeExistMap[s] = true
	}
	sort.Slice(c.sortedKeys, func(i, j int) bool {
		return c.sortedKeys[i] < c.sortedKeys[j]
	})
}
