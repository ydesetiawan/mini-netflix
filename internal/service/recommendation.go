package service

import (
	"context"
	"database/sql"

	"github.com/redis/go-redis/v9"
	rdb "github.com/ydesetiawan/mini-netflix/internal/redis"
	"github.com/ydesetiawan/mini-netflix/internal/model"
)

type RecommendationService struct {
	db    *sql.DB
	redis *redis.Client
}

func NewRecommendationService(db *sql.DB, rdb *redis.Client) *RecommendationService {
	return &RecommendationService{db: db, redis: rdb}
}

// MostPopular returns top N content by view count from Redis sorted set.
// Falls back to Postgres if Redis is empty.
func (s *RecommendationService) MostPopular(ctx context.Context, limit int) ([]model.Content, error) {
	ids, err := s.redis.ZRevRange(ctx, rdb.KeyPopularContent, 0, int64(limit-1)).Result()
	if err != nil || len(ids) == 0 {
		return s.popularFromDB(ctx, limit)
	}
	return s.fetchByIDs(ctx, ids)
}

// MostRecommended returns top N content by recommendation score (rating * log(count)).
// Falls back to Postgres if Redis is empty.
func (s *RecommendationService) MostRecommended(ctx context.Context, limit int) ([]model.Content, error) {
	ids, err := s.redis.ZRevRange(ctx, rdb.KeyRecommendedContent, 0, int64(limit-1)).Result()
	if err != nil || len(ids) == 0 {
		return s.recommendedFromDB(ctx, limit)
	}
	return s.fetchByIDs(ctx, ids)
}

func (s *RecommendationService) popularFromDB(ctx context.Context, limit int) ([]model.Content, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, description, type, release_year, thumbnail_url, view_count, rating_avg, rating_count
		FROM content ORDER BY view_count DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContentRows(rows)
}

func (s *RecommendationService) recommendedFromDB(ctx context.Context, limit int) ([]model.Content, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, description, type, release_year, thumbnail_url, view_count, rating_avg, rating_count
		FROM content
		WHERE rating_count > 0
		ORDER BY rating_avg * LN(rating_count + 1) DESC
		LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContentRows(rows)
}

func (s *RecommendationService) fetchByIDs(ctx context.Context, ids []string) ([]model.Content, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// Build ordered query preserving Redis rank order
	query := `
		SELECT id, title, description, type, release_year, thumbnail_url, view_count, rating_avg, rating_count
		FROM content WHERE id = ANY($1)`

	rows, err := s.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := map[string]model.Content{}
	for rows.Next() {
		var c model.Content
		rows.Scan(&c.ID, &c.Title, &c.Description, &c.Type, &c.ReleaseYear,
			&c.ThumbnailURL, &c.ViewCount, &c.RatingAvg, &c.RatingCount)
		byID[c.ID] = c
	}

	// Return in Redis rank order
	result := make([]model.Content, 0, len(ids))
	for _, id := range ids {
		if c, ok := byID[id]; ok {
			result = append(result, c)
		}
	}
	return result, nil
}

func scanContentRows(rows *sql.Rows) ([]model.Content, error) {
	var result []model.Content
	for rows.Next() {
		var c model.Content
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Type, &c.ReleaseYear,
			&c.ThumbnailURL, &c.ViewCount, &c.RatingAvg, &c.RatingCount); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
