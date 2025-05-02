package lb

import (
	"loadbalancer/errors"
	"log"
	"net/http"
	"sync"
)

type LoadBalancer struct {
	backends []*Backend
	current  uint64
	mux      sync.RWMutex
}

func NewLoadBalancer(serverURLs []string) (*LoadBalancer, error) {
	var backends []*Backend
	for _, serverURL := range serverURLs {
		backend, err := NewBackend(serverURL)
		if err != nil {
			return nil, err
		}
		backends = append(backends, backend)
	}

	return &LoadBalancer{
		backends: backends,
		current:  0,
	}, nil
}

func (lb *LoadBalancer) GetNextBackend() *Backend {
	lb.mux.Lock()
	defer lb.mux.Unlock()

	next := int(lb.current+1) % len(lb.backends)
	for i := 0; i < len(lb.backends); i++ {
		idx := (next + i) % len(lb.backends)
		if lb.backends[idx].IsAlive() {
			lb.current = uint64(idx)
			return lb.backends[idx]
		}
	}

	return nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.GetNextBackend()
	if backend != nil {
		log.Printf("Forwarding request to %s", backend.URL.String())
		backend.ReverseProxy.ServeHTTP(w, r)
		return
	}

	log.Println("No available backends")
	errors.JSONError(w, "No available backends", http.StatusServiceUnavailable)
}
