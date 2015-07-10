package loadbalancer

import (
	"sync/atomic"

	"github.com/go-kit/kit/endpoint"
)

// RoundRobin returns a load balancer that yields endpoints in sequence.
func RoundRobin(p Publisher) LoadBalancer {
	return &roundRobin{
		p: p,
	}
}

type roundRobin struct {
	p     Publisher
	count uint64
}

func (r *roundRobin) Count() int {
	return len(r.p.Endpoints())
}

func (r *roundRobin) Get() (endpoint.Endpoint, error) {
	endpoints := r.p.Endpoints()
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpointsAvailable
	}
	count := atomic.AddUint64(&r.count, 1)
	return endpoints[(count-1)%uint64(len(endpoints))], nil
}
