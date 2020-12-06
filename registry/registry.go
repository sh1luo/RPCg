package registry

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
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
	mu sync.RWMutex

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
	if r.timeout == 0 {
		r.servers[addr] = info
	}

	if info == "" {
		r.servers[addr] = "start=" + strconv.Itoa(int(time.Now().UnixNano()))
		return
	}

	// TODO:use parseInfo function instead
	v, _ := url.ParseQuery(info)
	if ww := v.Get("weight"); ww != "" {
		r.servers[addr] = fmt.Sprintf("weight=%s&start=%d", ww, int(time.Now().UnixNano()))
	}

	log.Println("put server完成:", r.servers)
}

func (r *RpcgRegistry) aliveServers() map[string]string {
	// Quickly release the lock when the servers is obtained
	r.mu.RLock()
	ss := r.servers
	r.mu.RUnlock()

	alive := make(map[string]string)
	// TODO:get the start value quickly?
	for addr, infos := range ss {
		if r.timeout == 0 {
			alive[addr] = infos
			continue
		}

		s := getStart(infos)
		if s > 0 && int64(s)+r.timeout.Nanoseconds() > time.Now().UnixNano() {
			alive[addr] = infos
			continue
		}
		delete(r.servers, addr)
	}

	return alive
}

// ServeHTTP Runs at /_rpcg_/registry
// // to keep it simple, all servers information is in req.Header
func (r *RpcgRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		as := r.aliveServers()
		for addr, info := range as {
			w.Header().Add("X-RPCg-Servers", addr)
			w.Header().Add("X-RPCg-Infos", info)
		}

	case "POST":
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

type payload struct {
	weight int
	start  uint

	// add ...
}

func parseInfo(info string, ptr interface{}) error {
	infoStr, err := url.ParseQuery(info)
	if err != nil {
		return err
	}

	// map key is the struct name
	fields := make(map[string]reflect.Value)
	v := reflect.ValueOf(ptr).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i)
		name := fieldInfo.Name
		fields[name] = v.Field(i)
	}

	for key, slice := range infoStr {
		f := fields[key]
		if !f.IsValid() {
			continue
		}
		for _, value := range slice {
			if err := populate(f, value); err != nil {
				return fmt.Errorf("%s : %v", key, err)
			}
		}
	}
	return nil
}

func populate(v reflect.Value, value string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(value)

	case reflect.Int:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(b)

	default:
		return fmt.Errorf("unsupported kind %s", v.Type())
	}

	return nil
}

func getStart(infos string) int {
	v, _ := url.ParseQuery(infos)
	vv := v.Get("start")
	t, _ := strconv.Atoi(vv)

	return t
}
