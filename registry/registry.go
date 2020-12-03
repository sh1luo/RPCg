package registry

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
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
	if info == "" {
		r.servers[addr] = "start=" + strconv.Itoa(int(time.Now().UnixNano()))
		return
	}

	if v, err := url.ParseQuery(info); err == nil {
		ww := v.Get("weight")
		w, err := strconv.Atoi(ww)
		if err != nil {
			log.Println(w, err)
		}
		r.servers[addr] = fmt.Sprintf("start=%s&weight=%d", time.Now().String(), w)
	}
}

func (r *RpcgRegistry) aliveServers() map[string]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	alive := make(map[string]string)
	for addr, info := range r.servers {
		if r.timeout == 0 {
			alive[addr] = info
			continue
		}
		if v, err := url.ParseQuery(info); err == nil {
			vv := v.Get("start")
			if t, err := strconv.Atoi(vv); err == nil {
				if t-int(time.Now().UnixNano()) > 0 {
					alive[addr] = info
				} else {
					delete(r.servers, addr)
				}
			}
		}
	}
	return alive
}

// Runs at /_rpcg_/registry
func (r *RpcgRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		// keep it simple, server is in req.Header
		as := r.aliveServers()
		fmt.Println("aliveServers:", as)
		for addr, info := range as {
			w.Header().Add("X-RPCg-Servers", addr)
			w.Header().Add("X-RPCg-Infos", info)
		}
	case "POST":
		// keep it simple, server is in req.Header
		addr := req.Header.Get("X-RPCg-Server-Addr")
		info := req.Header.Get("X-RPCg-Server-Info")
		if addr == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Println("[INFO]:Server alive:", addr, info)
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
func Heartbeat(registry, addr, info string, duration time.Duration) {
	if duration == 0 {
		// make sure there is enough time to send heart beat
		// before it's removed from registry
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr, info)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr, info)
		}
	}()
}

func sendHeartbeat(registry, addr, info string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-RPCg-Server-Addr", addr)
	req.Header.Set("X-RPCg-Server-Info", info)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat err:", err)
		return err
	}
	return nil
}
