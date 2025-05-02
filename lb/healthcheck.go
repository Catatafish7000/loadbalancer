package lb

import (
	"log"
	"net/http"
	"time"
)

func (lb *LoadBalancer) HealthCheck(interval time.Duration) {
	for _, b := range lb.backends {
		go func(backend *Backend) {
			t := time.NewTicker(interval)
			for {
				select {
				case <-t.C:
					alive := isBackendAlive(backend.URL.String())
					backend.SetAlive(alive)
					status := "up"
					if !alive {
						status = "down"
					}
					log.Printf("%s [%s]\n", backend.URL.String(), status)
				}
			}
		}(b)
	}
}

func isBackendAlive(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error checking backend %s: %v", url, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500
}
