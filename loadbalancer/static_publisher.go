package loadbalancer

import (
	"sync"

	"github.com/go-kit/kit/endpoint"
)

// StaticPublisher holds a static set of endpoints.
type StaticPublisher struct {
	mu      sync.Mutex
	current []endpoint.Endpoint
}

// NewStaticPublisher returns a publisher that yields a static set of
// endpoints which can be completely replaced.
func NewStaticPublisher(endpoints []endpoint.Endpoint) *StaticPublisher {
	return &StaticPublisher{
		current: endpoints,
	}
}

// Endpoints implements Publisher.Endpoints.
func (p *StaticPublisher) Endpoints() []endpoint.Endpoint {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

// Replace replaces the endpoints and notifies all subscribers.
// The caller should not mutate the endpoints slice after calling
// this function.
func (p *StaticPublisher) Replace(endpoints []endpoint.Endpoint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = endpoints
}
