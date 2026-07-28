CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_create_id UUID NOT NULL,
    time_start TIMESTAMP NOT NULL,
    create_time TIMESTAMP NOT NULL, 
    user_count INT NOT NULL DEFAULT 0,
) ORDER BY (time_start);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    event_id UUID,
    join_time TIMESTAMP,
);
