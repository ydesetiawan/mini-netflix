package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	es "github.com/ydesetiawan/mini-netflix/internal/elasticsearch"
	rdb "github.com/ydesetiawan/mini-netflix/internal/redis"
)

type SearchService struct {
	es    *es.Client
	redis *redis.Client
}

func NewSearchService(es *es.Client, rdb *redis.Client) *SearchService {
	return &SearchService{es: es, redis: rdb}
}

type SearchResult struct {
	Total int              `json:"total"`
	Hits  []map[string]any `json:"hits"`
}

// Autocomplete returns title suggestions for a prefix (fast, cached 1 min).
func (s *SearchService) Autocomplete(ctx context.Context, prefix string) ([]string, error) {
	cacheKey := fmt.Sprintf(rdb.KeySearchCache, "autocomplete:"+prefix)

	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var result []string
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	}

	suggestions, err := s.es.Autocomplete(ctx, prefix)
	if err != nil {
		return nil, err
	}

	if b, err := json.Marshal(suggestions); err == nil {
		s.redis.Set(ctx, cacheKey, b, 1*time.Minute)
	}
	return suggestions, nil
}

// Search performs full-text fuzzy search with Redis caching (2 min TTL).
func (s *SearchService) Search(ctx context.Context, q string, page, size int) (*SearchResult, error) {
	from := page * size
	cacheKey := fmt.Sprintf(rdb.KeySearchCache, fmt.Sprintf("%s:%d:%d", q, page, size))

	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var result SearchResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	hits, total, err := s.es.Search(ctx, q, from, size)
	if err != nil {
		return nil, err
	}

	result := &SearchResult{Total: total, Hits: hits}

	if b, err := json.Marshal(result); err == nil {
		s.redis.Set(ctx, cacheKey, b, 2*time.Minute)
	}
	return result, nil
}
