CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_create_id UUID NOT NULL,
    enemy_side_leader UUID,
    time_start TIMESTAMP NOT NULL,
    time_finish TIMESTAMP,
    create_time TIMESTAMP NOT NULL,
    user_count INT NOT NULL DEFAULT 0,
    event_team_winner TEXT,
    event_team_loser TEXT
) ORDER BY (time_start);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    event_id UUID,
    clan_id TEXT,
    team_id TEXT,
    role TEXT NOT NULL,
    six_clan_members BOOLEAN NOT NULL DEFAULT false,
    join_time TIMESTAMP
);

CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL,
    user_create_id UUID NOT NULL,
    enemy_side_leader UUID NOT NULL,
    team1_id UUID,
    team2_id UUID,
    map_name TEXT NOT NULL,
    game_team_winner_id TEXT,
    game_team_loser_id TEXT,
    time_start TIMESTAMP NOT NULL,
    time_finish TIMESTAMP
);

CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL,
    side_leader_id UUID NOT NULL,
    is_confirmed BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS team_members (
    team_id UUID NOT NULL,
    user_id UUID NOT NULL,
    clan_id TEXT,
    role TEXT NOT NULL,
    PRIMARY KEY (team_id, user_id)
);

CREATE TABLE IF NOT EXISTS game_user_stats (
    id UUID PRIMARY KEY,
    game_id UUID NOT NULL,
    user_id UUID NOT NULL,
    clan_id TEXT,
    team_id TEXT,
    role TEXT NOT NULL,
    kills BIGINT NOT NULL DEFAULT 0,
    deaths BIGINT NOT NULL DEFAULT 0,
    points BIGINT NOT NULL DEFAULT 0
);
