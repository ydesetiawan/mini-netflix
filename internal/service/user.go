package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/ydesetiawan/mini-netflix/internal/model"
)

type UserService struct {
	db        *sql.DB
	jwtSecret string
	jwtExpiry int
}

func NewUserService(db *sql.DB, jwtSecret string, jwtExpiry int) *UserService {
	return &UserService{db: db, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

func (s *UserService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var u model.User
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO users (email, username, password)
		VALUES ($1, $2, $3)
		RETURNING id, email, username, created_at`,
		req.Email, req.Username, string(hashed),
	).Scan(&u.ID, &u.Email, &u.Username, &u.CreatedAt)
	if err != nil {
		return nil, err
	}

	token, err := s.generateToken(u.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{Token: token, User: u}, nil
}

func (s *UserService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	var u model.User
	err := s.db.QueryRowContext(ctx, `
		SELECT id, email, username, password, created_at FROM users WHERE email = $1`, req.Email,
	).Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateToken(u.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{Token: token, User: u}, nil
}

func (s *UserService) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Duration(s.jwtExpiry) * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *UserService) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	userID, _ := claims["sub"].(string)
	return userID, nil
}

func (s *UserService) AddToWatchlist(ctx context.Context, userID, contentID string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO watchlist (user_id, content_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, contentID,
	)
	return err
}

func (s *UserService) GetWatchlist(ctx context.Context, userID string) ([]model.Content, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.title, c.description, c.type, c.release_year,
		       c.thumbnail_url, c.view_count, c.rating_avg, c.rating_count
		FROM content c
		JOIN watchlist w ON w.content_id = c.id
		WHERE w.user_id = $1
		ORDER BY w.added_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContentRows(rows)
}
