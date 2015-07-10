package loadbalancer

import (
	"math/rand"

	"github.com/go-kit/kit/endpoint"
)

// Random returns a load balancer that yields random endpoints.
func Random(p Publisher) LoadBalancer {
	return random{p}
}

type random struct {
	p Publisher
}

func (r random) Count() int { return len(r.p.Endpoints()) }

func (r random) Get() (endpoint.Endpoint, error) {
	endpoints := r.p.Endpoints()
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpointsAvailable
	}
	return endpoints[rand.Intn(len(endpoints))], nil
}
