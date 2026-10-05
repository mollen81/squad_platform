CREATE TABLE event (
    id UUID PRIMARY KEY NOT NULL,
    name VARCHAR(255),
    creator_user_id UUID NOT NULL,
    target_game_count INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    time_start TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_event_creator_user_id ON event(creator_user_id);
CREATE INDEX idx_event_time_start ON event(time_start);


CREATE TABLE event_side (
    id UUID PRIMARY KEY NOT NULL,
    event_id UUID NOT NULL REFERENCES event(id) ON DELETE CASCADE,
    name VARCHAR(255),
    leader_user_id UUID NOT NULL,
    is_ready BOOLEAN DEFAULT FALSE NOT NULL,
    ready_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_event_side_event_id ON event_side(event_id);


CREATE TABLE event_participant (
    id UUID PRIMARY KEY NOT NULL,
    event_id UUID NOT NULL,
    side_id UUID NOT NULL,
    user_id UUID NOT NULL,
    clan_id UUID,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT uq_event_participant_event_user UNIQUE (event_id, user_id)
);

CREATE INDEX idx_event_participant_event_id ON event_participant(event_id);
CREATE INDEX idx_event_participant_side_id ON event_participant(side_id);


