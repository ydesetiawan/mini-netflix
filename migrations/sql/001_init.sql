-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- for trigram similarity search

-- Users
CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email      VARCHAR(255) UNIQUE NOT NULL,
    username   VARCHAR(100) UNIQUE NOT NULL,
    password   VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Genres
CREATE TABLE genres (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

-- Content (movies & series)
CREATE TABLE content (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title        VARCHAR(500) NOT NULL,
    description  TEXT,
    type         VARCHAR(20) NOT NULL CHECK (type IN ('movie', 'series')),
    release_year INT,
    duration_min INT,          -- for movies
    total_episodes INT,        -- for series
    thumbnail_url TEXT,
    video_url    TEXT,         -- CDN URL
    view_count   BIGINT DEFAULT 0,
    rating_avg   NUMERIC(3,1) DEFAULT 0,
    rating_count INT DEFAULT 0,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Content <-> Genres
CREATE TABLE content_genres (
    content_id UUID REFERENCES content(id) ON DELETE CASCADE,
    genre_id   INT  REFERENCES genres(id)  ON DELETE CASCADE,
    PRIMARY KEY (content_id, genre_id)
);

-- Watch events (for popularity tracking)
CREATE TABLE watch_events (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    content_id UUID REFERENCES content(id) ON DELETE CASCADE,
    watched_at TIMESTAMPTZ DEFAULT NOW(),
    progress   INT DEFAULT 0  -- seconds watched
);

-- Ratings
CREATE TABLE ratings (
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    content_id UUID REFERENCES content(id) ON DELETE CASCADE,
    score      INT NOT NULL CHECK (score BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, content_id)
);

-- Watchlist
CREATE TABLE watchlist (
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    content_id UUID REFERENCES content(id) ON DELETE CASCADE,
    added_at   TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, content_id)
);

-- Indexes
CREATE INDEX idx_content_title_trgm ON content USING GIN (title gin_trgm_ops);
CREATE INDEX idx_content_view_count  ON content (view_count DESC);
CREATE INDEX idx_content_rating_avg  ON content (rating_avg DESC);
CREATE INDEX idx_watch_events_content ON watch_events (content_id);
CREATE INDEX idx_watch_events_user    ON watch_events (user_id);
