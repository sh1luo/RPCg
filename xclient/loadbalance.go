package xclient

import (
	"net/url"
	"rpcg/utils"
	"strconv"
)

type Balancer interface {
	Pick() string
	UpdateAllServers(servers map[string]string)
}

func newBalancer(selectMode SelectMode, servers map[string]string) Balancer {
	switch selectMode {
	case Random:
		return newRandomBalancer(servers)
	case RoundRobin:
		return newRoundRobinSelector(servers)
	case WeightedRoundRobin:
		return newWeightedRoundRobinSelector(servers)
	case ConsistentHash:
		return newConsistentHashBalancer()
	default:
		return newRandomBalancer(servers)
	}
}

type randomBalancer struct {
	servers []string
}

func newRandomBalancer(servers map[string]string) Balancer {
	b := &randomBalancer{servers: make([]string, len(servers), len(servers))}
	for _, s := range servers {
		b.servers = append(b.servers, s)
	}
	return b
}

func (r randomBalancer) Pick() string {
	return r.servers[utils.Intn(len(r.servers))]
}

func (r randomBalancer) UpdateAllServers(servers map[string]string) {
	ss := make([]string, 0, len(servers))
	for k := range servers {
		ss = append(ss, k)
	}

	r.servers = ss
}

// roundRobinSelector selects servers with roundrobin.
type roundRobinBalancer struct {
	servers []string
	i       int
}

func newRoundRobinSelector(servers map[string]string) Balancer {
	ss := make([]string, 0, len(servers))
	for k := range servers {
		ss = append(ss, k)
	}

	return &roundRobinBalancer{servers: ss}
}

func (s *roundRobinBalancer) Pick() string {
	ss := s.servers
	if len(ss) == 0 {
		return ""
	}
	i := s.i
	i = i % len(ss)
	s.i = i + 1

	return ss[i]
}

func (s *roundRobinBalancer) UpdateAllServers(servers map[string]string) {
	ss := make([]string, 0, len(servers))
	for k := range servers {
		ss = append(ss, k)
	}

	s.servers = ss
}

// weightedRoundRobinSelector selects servers with weighted.
type weightedRoundRobinSelector struct {
	servers []*Weighted
}

func newWeightedRoundRobinSelector(servers map[string]string) Balancer {
	ss := createWeighted(servers)
	return &weightedRoundRobinSelector{servers: ss}
}

func (s *weightedRoundRobinSelector) Pick() string {
	ss := s.servers
	if len(ss) == 0 {
		return ""
	}
	w := nextWeighted(ss)
	if w == nil {
		return ""
	}
	return w.Server
}

func (s *weightedRoundRobinSelector) UpdateAllServers(servers map[string]string) {
	ss := createWeighted(servers)
	s.servers = ss
}

func createWeighted(servers map[string]string) []*Weighted {
	ss := make([]*Weighted, 0, len(servers))
	for k, metadata := range servers {
		w := &Weighted{Server: k, Weight: 1}

		if v, err := url.ParseQuery(metadata); err == nil {
			ww := v.Get("weight")
			if ww != "" {
				if weight, err := strconv.Atoi(ww); err == nil {
					w.Weight = weight
				}
			}
		}

		ss = append(ss, w)
	}

	return ss
}