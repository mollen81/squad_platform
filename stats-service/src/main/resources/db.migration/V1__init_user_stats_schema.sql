CREATE TABLE user_stats (
    user_id UUID PRIMARY KEY,
    steam_id VARCHAR(20) NOT NULL UNIQUE,
    elo_rating INT DEFAULT 1000,
    matches_played INT DEFAULT 0,
    total_playtime_hours INT DEFAULT 0,
    kills INT DEFAULT 0,
    deaths INT DEFAULT 0,
    revives INT DEFAULT 0,
    favorite_role VARCHAR(50),
    last_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
)