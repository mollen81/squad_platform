CREATE TABLE clan_applications(
    id UUID PRIMARY KEY,
    userId UUID NOT NULL,
    clan_id UUID NOT NULL REFERENCES clans(id) ON DELETE CASCADE,
    social_link VARCHAR(255) NOT NULL,
    experience_text TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_clan_user_status UNIQUE (clan_id, user_id, status)
)

CREATE INDEX idx_clan_applications_clan_id_status ON clan_applications(clan_id, status);
CREATE INDEX idx_clan_applications_user_id ON clan_applications(user_id);