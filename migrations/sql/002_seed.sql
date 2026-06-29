-- Genres
INSERT INTO genres (name) VALUES
    ('Action'), ('Drama'), ('Comedy'), ('Sci-Fi'), ('Horror'),
    ('Romance'), ('Thriller'), ('Animation'), ('Documentary'), ('Fantasy')
ON CONFLICT DO NOTHING;

-- Sample content
INSERT INTO content (id, title, description, type, release_year, duration_min, thumbnail_url, video_url, view_count, rating_avg, rating_count)
VALUES
    ('a1b2c3d4-0001-0001-0001-000000000001', 'The Dark Knight', 'Batman faces the Joker in Gotham City.', 'movie', 2008, 152, 'https://placeholder.co/300x450', 'https://cdn.example.com/dark-knight.m3u8', 9800, 9.2, 420),
    ('a1b2c3d4-0002-0002-0002-000000000002', 'Inception', 'A thief enters dreams to plant an idea.', 'movie', 2010, 148, 'https://placeholder.co/300x450', 'https://cdn.example.com/inception.m3u8', 8700, 8.8, 380),
    ('a1b2c3d4-0003-0003-0003-000000000003', 'Breaking Bad', 'A chemistry teacher turns to making meth.', 'series', 2008, NULL, 'https://placeholder.co/300x450', 'https://cdn.example.com/bb.m3u8', 15000, 9.5, 900),
    ('a1b2c3d4-0004-0004-0004-000000000004', 'Interstellar', 'Astronauts travel through a wormhole near Saturn.', 'movie', 2014, 169, 'https://placeholder.co/300x450', 'https://cdn.example.com/interstellar.m3u8', 7200, 8.6, 310),
    ('a1b2c3d4-0005-0005-0005-000000000005', 'Stranger Things', 'Kids in a small town face supernatural forces.', 'series', 2016, NULL, 'https://placeholder.co/300x450', 'https://cdn.example.com/st.m3u8', 12000, 8.7, 650)
ON CONFLICT DO NOTHING;

-- Content <-> Genres
INSERT INTO content_genres (content_id, genre_id) VALUES
    ('a1b2c3d4-0001-0001-0001-000000000001', 1), -- Dark Knight: Action
    ('a1b2c3d4-0001-0001-0001-000000000001', 7), -- Thriller
    ('a1b2c3d4-0002-0002-0002-000000000002', 4), -- Inception: Sci-Fi
    ('a1b2c3d4-0002-0002-0002-000000000002', 7), -- Thriller
    ('a1b2c3d4-0003-0003-0003-000000000003', 2), -- Breaking Bad: Drama
    ('a1b2c3d4-0003-0003-0003-000000000003', 7), -- Thriller
    ('a1b2c3d4-0004-0004-0004-000000000004', 4), -- Interstellar: Sci-Fi
    ('a1b2c3d4-0004-0004-0004-000000000004', 2), -- Drama
    ('a1b2c3d4-0005-0005-0005-000000000005', 5), -- Stranger Things: Horror
    ('a1b2c3d4-0005-0005-0005-000000000005', 4)  -- Sci-Fi
ON CONFLICT DO NOTHING;
