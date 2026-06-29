package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ydesetiawan/mini-netflix/internal/config"
	"github.com/ydesetiawan/mini-netflix/internal/db"
	"github.com/ydesetiawan/mini-netflix/internal/elasticsearch"
	"github.com/ydesetiawan/mini-netflix/internal/handler"
	rdb "github.com/ydesetiawan/mini-netflix/internal/redis"
	"github.com/ydesetiawan/mini-netflix/internal/router"
	"github.com/ydesetiawan/mini-netflix/internal/service"
)

func main() {
	cfg := config.Load()

	// Connect to Postgres
	postgres, err := db.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer postgres.Close()
	log.Println("✓ PostgreSQL connected")

	// Connect to Redis
	redisClient, err := rdb.NewClient(cfg)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("✓ Redis connected")

	// Connect to Elasticsearch
	esClient := elasticsearch.NewClient(cfg)
	// Create ES index (idempotent - will return 400 if exists, that's ok)
	if err := esClient.CreateIndex(context.Background()); err != nil {
		log.Printf("ES index (may already exist): %v", err)
	} else {
		log.Println("✓ Elasticsearch index ready")
	}

	// Wire up services
	contentSvc := service.NewContentService(postgres, redisClient, esClient)
	searchSvc := service.NewSearchService(esClient, redisClient)
	recSvc := service.NewRecommendationService(postgres, redisClient)
	userSvc := service.NewUserService(postgres, cfg.JWTSecret, cfg.JWTExpiryHours)

	// Wire up handlers
	contentH := handler.NewContentHandler(contentSvc)
	searchH := handler.NewSearchHandler(searchSvc)
	recH := handler.NewRecommendationHandler(recSvc)
	userH := handler.NewUserHandler(userSvc)

	// Setup router
	r := router.Setup(userSvc, contentH, searchH, recH, userH)

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 Mini Netflix API running on :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("server stopped")
}
