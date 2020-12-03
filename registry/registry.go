package registry

import (
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

// RpcgRegistry is a simple register center, provide following functions.
// add a server and receive heartbeat to keep it alive.
// returns all alive servers and delete dead servers sync simultaneously.
type RpcgRegistry struct {
	timeout time.Duration

	// protect following
	mu sync.Mutex

	// map's values are expected to separated byampersands or semicolons,
	// e.g."weight=xxx&start=xxx"
	servers map[string]string
}

const (
	defaultPath    = "/_rpcg_/registry"
	defaultTimeout = time.Minute * 5
)

// New create a registry instance with timeout setting
func New(timeout time.Duration) *RpcgRegistry {
	return &RpcgRegistry{
		servers: make(map[string]string),
		timeout: timeout,
	}
}

var DefaultGeeRegister = New(defaultTimeout)

func (r *RpcgRegistry) putServer(addr, info string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exist := r.servers[addr]; exist {
		r.servers[addr] = info
		if v, err := url.ParseQuery(info); err == nil && v.Get("start") == "" {
			r.servers[addr] += "&start=" + time.Now().String()
		}
	}
}

func (r *RpcgRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, info := range r.servers {
		if v, err := url.ParseQuery(info); err == nil {
			s := v.Get("start")
			if start, err := time.Parse("2006-01-02 15:04:05", s); err == nil && s != "" {
				if r.timeout == 0 || start.Add(r.timeout).After(time.Now()) {
					alive = append(alive, addr)
				} else {
					delete(r.servers, addr)
				}
			}
		}
	}
	sort.Strings(alive)
	return alive
}

// Runs at /_rpcg_/registry
func (r *RpcgRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		// keep it simple, server is in req.Header
		w.Header().Set("X-RPCg-Servers", strings.Join(r.aliveServers(), ","))
	case "POST":
		// keep it simple, server is in req.Header
		addr := req.Header.Get("X-RPCg-Server-Addr")
		info := req.Header.Get("X-RPCg-Server-Info")
		if addr == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.putServer(addr, info)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// HandleHTTP registers an HTTP handler for RpcgRegistry messages on registryPath
func (r *RpcgRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path:", registryPath)
}

func HandleHTTP() {
	DefaultGeeRegister.HandleHTTP(defaultPath)
}

// Heartbeat send a heartbeat message every once in a while
// it's a helper function for a server to register or send heartbeat
func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		// make sure there is enough time to send heart beat
		// before it's removed from registry
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-RPCg-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat err:", err)
		return err
	}
	return nil
}
