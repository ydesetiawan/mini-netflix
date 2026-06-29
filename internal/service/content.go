package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	rdb "github.com/ydesetiawan/mini-netflix/internal/redis"
	es "github.com/ydesetiawan/mini-netflix/internal/elasticsearch"
	"github.com/ydesetiawan/mini-netflix/internal/model"
)

type ContentService struct {
	db    *sql.DB
	redis *redis.Client
	es    *es.Client
}

func NewContentService(db *sql.DB, rdb *redis.Client, es *es.Client) *ContentService {
	return &ContentService{db: db, redis: rdb, es: es}
}

func (s *ContentService) Create(ctx context.Context, req model.CreateContentRequest) (*model.Content, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var c model.Content
	err = tx.QueryRowContext(ctx, `
		INSERT INTO content (title, description, type, release_year, duration_min, total_episodes, thumbnail_url, video_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, title, description, type, release_year, duration_min, total_episodes,
		          thumbnail_url, video_url, view_count, rating_avg, rating_count, created_at`,
		req.Title, req.Description, req.Type, req.ReleaseYear,
		req.DurationMin, req.TotalEpisodes, req.ThumbnailURL, req.VideoURL,
	).Scan(
		&c.ID, &c.Title, &c.Description, &c.Type, &c.ReleaseYear,
		&c.DurationMin, &c.TotalEpisodes, &c.ThumbnailURL, &c.VideoURL,
		&c.ViewCount, &c.RatingAvg, &c.RatingCount, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	for _, gID := range req.GenreIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO content_genres (content_id, genre_id) VALUES ($1, $2)`, c.ID, gID,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	c.Genres = s.genreNames(ctx, req.GenreIDs)

	// Index in Elasticsearch
	go s.indexInES(context.Background(), &c)

	return &c, nil
}

func (s *ContentService) GetByID(ctx context.Context, id string) (*model.Content, error) {
	cacheKey := fmt.Sprintf(rdb.KeyContentCache, id)

	// Try Redis cache first
	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var c model.Content
		if json.Unmarshal([]byte(cached), &c) == nil {
			return &c, nil
		}
	}

	var c model.Content
	err := s.db.QueryRowContext(ctx, `
		SELECT id, title, description, type, release_year, duration_min, total_episodes,
		       thumbnail_url, video_url, view_count, rating_avg, rating_count, created_at
		FROM content WHERE id = $1`, id,
	).Scan(
		&c.ID, &c.Title, &c.Description, &c.Type, &c.ReleaseYear,
		&c.DurationMin, &c.TotalEpisodes, &c.ThumbnailURL, &c.VideoURL,
		&c.ViewCount, &c.RatingAvg, &c.RatingCount, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	c.Genres = s.fetchGenres(ctx, id)

	// Cache for 5 minutes
	if b, err := json.Marshal(c); err == nil {
		s.redis.Set(ctx, cacheKey, b, 5*time.Minute)
	}

	return &c, nil
}

// RecordWatch increments view count in DB and Redis sorted set for popularity.
func (s *ContentService) RecordWatch(ctx context.Context, ev model.WatchEvent) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO watch_events (user_id, content_id, progress) VALUES ($1, $2, $3)`,
		ev.UserID, ev.ContentID, ev.Progress,
	)
	if err != nil {
		return err
	}

	// Increment in DB
	s.db.ExecContext(ctx, `UPDATE content SET view_count = view_count + 1 WHERE id = $1`, ev.ContentID)

	// Update Redis popularity sorted set (ZINCRBY adds score atomically)
	s.redis.ZIncrBy(ctx, rdb.KeyPopularContent, 1, ev.ContentID)

	// Invalidate content cache
	s.redis.Del(ctx, fmt.Sprintf(rdb.KeyContentCache, ev.ContentID))

	return nil
}

// Rate records a user rating and updates the running average.
func (s *ContentService) Rate(ctx context.Context, userID, contentID string, score int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Upsert rating
	_, err = tx.ExecContext(ctx, `
		INSERT INTO ratings (user_id, content_id, score) VALUES ($1, $2, $3)
		ON CONFLICT (user_id, content_id) DO UPDATE SET score = $3, created_at = NOW()`,
		userID, contentID, score,
	)
	if err != nil {
		return err
	}

	// Recompute running average
	_, err = tx.ExecContext(ctx, `
		UPDATE content SET
			rating_avg   = (SELECT AVG(score) FROM ratings WHERE content_id = $1),
			rating_count = (SELECT COUNT(*)  FROM ratings WHERE content_id = $1)
		WHERE id = $1`, contentID,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Update recommendation score: avg_rating * log(rating_count + 1)
	var avg float64
	var cnt int
	s.db.QueryRowContext(ctx, `SELECT rating_avg, rating_count FROM content WHERE id=$1`, contentID).Scan(&avg, &cnt)
	recScore := avg * logScale(cnt)
	s.redis.ZAdd(ctx, rdb.KeyRecommendedContent, redis.Z{Score: recScore, Member: contentID})

	// Invalidate cache
	s.redis.Del(ctx, fmt.Sprintf(rdb.KeyContentCache, contentID))
	return nil
}

func (s *ContentService) genreNames(ctx context.Context, ids []int) []string {
	if len(ids) == 0 {
		return nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT name FROM genres WHERE id = ANY($1)`, intSliceToArray(ids))
	if err != nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	return names
}

func (s *ContentService) fetchGenres(ctx context.Context, contentID string) []string {
	rows, err := s.db.QueryContext(ctx, `
		SELECT g.name FROM genres g
		JOIN content_genres cg ON cg.genre_id = g.id
		WHERE cg.content_id = $1`, contentID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	return names
}

func (s *ContentService) indexInES(ctx context.Context, c *model.Content) {
	doc := map[string]any{
		"id":            c.ID,
		"title":         c.Title,
		"description":   c.Description,
		"type":          c.Type,
		"release_year":  c.ReleaseYear,
		"genres":        c.Genres,
		"view_count":    c.ViewCount,
		"rating_avg":    c.RatingAvg,
		"thumbnail_url": c.ThumbnailURL,
		"title_suggest": map[string]any{"input": []string{c.Title}},
	}
	s.es.IndexContent(ctx, doc)
}

func intSliceToArray(ids []int) interface{} {
	return ids
}

func logScale(n int) float64 {
	if n <= 0 {
		return 0
	}
	// ln(n+1) approximation using repeated halving
	x := float64(n + 1)
	result := 0.0
	for x > 1 {
		x /= 2.718281828
		result++
	}
	return result
}
