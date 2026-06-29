package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/ydesetiawan/mini-netflix/internal/config"
)

func NewClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

// Key constants
const (
	KeyPopularContent     = "popular:content"       // sorted set: content_id -> view_count
	KeyRecommendedContent = "recommended:content"    // sorted set: content_id -> score
	KeyContentCache       = "cache:content:%s"       // string: JSON content by ID
	KeySearchCache        = "cache:search:%s"        // string: JSON results by query
	KeyUserSession        = "session:user:%s"        // string: JWT by user ID
)
