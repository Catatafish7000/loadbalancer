package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"loadbalancer/config"
	"loadbalancer/lb"
	"loadbalancer/ratelimiter"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	balancer, err := lb.NewLoadBalancer(cfg.Backends)
	if err != nil {
		log.Fatalf("Failed to create balancer: %v", err)
	}

	limiter := ratelimiter.NewLimiter(
		cfg.RateLimit.StoragePath,
		cfg.RateLimit.Capacity,
		cfg.RateLimit.RatePerSec,
	)
	defer limiter.Close()

	handler := limiter.Middleware(balancer)
	balancer.HealthCheck(10 * time.Second)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	go func() {
		log.Printf("Server started on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
