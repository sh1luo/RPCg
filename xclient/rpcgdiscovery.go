package xclient

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

type RpcgRegistryDiscovery struct {
	*MultiServersDiscovery
	registry   string
	timeout    time.Duration
	lastUpdate time.Time
}

const defaultUpdateTimeout = time.Second

func (d *RpcgRegistryDiscovery) Update(servers map[string]string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d *RpcgRegistryDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	log.Println("rpc registry: refresh servers from registry", d.registry)
	resp, err := http.Get(d.registry)
	if err != nil {
		log.Println("rpc registry refresh err:", err)
		return err
	}

	servers := resp.Header.Values("X-RPCg-Servers")
	infos := resp.Header.Values("X-RPCg-Infos")
	if len(servers) != len(infos) {
		log.Printf("rpc registry get http header err:\n\tservers:%s\n\binfos:%s", servers, infos)
		return errors.New("rpc registry :get http header err")
	}

	d.servers = make(map[string]string, len(servers))
	for k, server := range servers {
		d.servers[server] = infos[k]
	}
	d.lastUpdate = time.Now()

	return nil
}

func (d *RpcgRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		fmt.Println("d.Refresh err:", err)
		return "", err
	}

	return d.MultiServersDiscovery.Get(mode)
}

func (d *RpcgRegistryDiscovery) GetAll() (map[string]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServersDiscovery.GetAll()
}

func NewRegistryDiscovery(registerAddr string, timeout time.Duration) *RpcgRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &RpcgRegistryDiscovery{
		MultiServersDiscovery: NewMultiServerDiscovery(make(map[string]string, 0)),
		registry:              registerAddr,
		timeout:               timeout,
	}
	return d
}
