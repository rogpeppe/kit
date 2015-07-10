package loadbalancer

import "github.com/go-kit/kit/endpoint"

// Publisher provides access to a set of endpoints
// that may change over time.
type Publisher interface {
	// Endpoints returns the current endpoints.
	// The caller should not mutate the returned slice.
	Endpoints() []endpoint.Endpoint
}
