package xclient

import (
	"errors"
	"math/rand"
	"rpcg/utils"
	"sync"
)

type Discovery interface {
	Refresh() error // refresh from remote registry
	Update(servers map[string]string) error
	Get(mode SelectMode) (string, error)
	GetAll() (map[string]string, error)
}

var _ Discovery = (*MultiServersDiscovery)(nil)

// MultiServersDiscovery is a discovery for multi servers without a registry center
// user provides the server addresses explicitly instead
type MultiServersDiscovery struct {
	r  *rand.Rand   // generate random number
	mu sync.RWMutex // protect following
	//servers []string
	servers map[string]string
	index   int // record the selected position for robin algorithm
}

// Refresh doesn't make sense for MultiServersDiscovery, so ignore it
func (d *MultiServersDiscovery) Refresh() error {
	return nil
}

// Update the servers of discovery dynamically if needed
func (d *MultiServersDiscovery) Update(servers map[string]string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

// Get a server according to mode
func (d *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}

	balancer := newBalancer(mode, d.servers)
	s := balancer.Pick()
	return s, nil
}

// returns all servers in discovery
func (d *MultiServersDiscovery) GetAll() (map[string]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	// return a copy of d.servers
	servers := make(map[string]string, len(d.servers))
	utils.CopyMap(servers, d.servers)
	return servers, nil
}

// NewMultiServerDiscovery creates a MultiServersDiscovery instance
func NewMultiServerDiscovery(servers map[string]string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
	}
	return d
}
