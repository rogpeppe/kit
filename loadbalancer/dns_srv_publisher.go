package loadbalancer

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/go-kit/kit/endpoint"
)

type DNSSRVPublisher struct {
	quit chan struct{}

	mu        sync.Mutex
	endpoints []endpoint.Endpoint
}

// NewDNSSRVPublisher returns a publisher that resolves the SRV name every ttl, and
func NewDNSSRVPublisher(name string, ttl time.Duration, makeEndpoint func(hostport string) endpoint.Endpoint) *DNSSRVPublisher {
	// TODO resolve addresses immediately so that there
	// is always a valid set of endpoints available?
	p := &DNSSRVPublisher{
		quit: make(chan struct{}),
	}
	// Acquire the mutex before calling p.loop so that
	// the creation of this type doesn't block but
	// the first call to Endpoints will block
	// until the initial endpoints are available.
	p.mu.Lock()
	go p.loop(name, ttl, makeEndpoint)
	return p
}

func (p *DNSSRVPublisher) Stop() {
	close(p.quit)
}

func (p *DNSSRVPublisher) Endpoints() []endpoint.Endpoint {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.endpoints
}

var newTicker = time.NewTicker

// loop runs the publisher's polling loop. It is called with p.mu locked.
func (p *DNSSRVPublisher) loop(name string, ttl time.Duration, makeEndpoint func(hostport string) endpoint.Endpoint) {
	addrs, err := resolve(name)
	if err == nil {
		p.endpoints = convert(addrs, makeEndpoint)
	}
	p.mu.Unlock()

	ticker := newTicker(ttl)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			newAddrs, err := resolve(name)
			if err == nil && !srvSliceEqual(newAddrs, addrs) {
				addrs = newAddrs
				endpoints := convert(addrs, makeEndpoint)
				p.mu.Lock()
				p.endpoints = endpoints
				p.mu.Unlock()
			}

		case <-p.quit:
			return
		}
	}
}

func srvSliceEqual(ss0, ss1 []*net.SRV) bool {
	if len(ss0) != len(ss1) {
		return false
	}
	for i, s0 := range ss0 {
		s1 := ss1[i]
		if s0.Target != s1.Target || s0.Port != s1.Port {
			return false
		}
	}
	return true
}

// resolve resolves the given name to its current addresses, and
// returns an Allow mocking in tests.
var resolve = func(name string) ([]*net.SRV, error) {
	_, addrs, err := net.LookupSRV("", "", name)
	sort.Sort(srvByTarget(addrs))
	return addrs, err
}

func convert(addrs []*net.SRV, makeEndpoint func(hostport string) endpoint.Endpoint) []endpoint.Endpoint {
	endpoints := make([]endpoint.Endpoint, len(addrs))
	for i, addr := range addrs {
		endpoints[i] = makeEndpoint(addr2hostport(addr))
	}
	return endpoints
}

func addr2hostport(addr *net.SRV) string {
	return net.JoinHostPort(addr.Target, fmt.Sprintf("%d", addr.Port))
}

type srvByTarget []*net.SRV

func (s srvByTarget) Less(i, j int) bool {
	s0 := s[i]
	s1 := s[j]
	if s0.Target != s1.Target {
		return s0.Target < s1.Target
	}
	return s0.Port < s1.Port
}

func (s srvByTarget) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s srvByTarget) Len() int {
	return len(s)
}
