CREATE TABLE clans (
   id UUID PRIMARY KEY,
   name VARCHAR(100) UNIQUE NOT NULL,
   tag VARCHAR(10) NOT NULL,
   description TEXT,
   requirements TEXT,
   avatar_url VARCHAR(255),
   is_recruiting BOOLEAN DEFAULT true,
   status VARCHAR(20) DEFAULT 'UNVERIFIED', -- UNVERIFIED, OFFICIAL
   total_elo INT DEFAULT 0,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE clan_members (
  id UUID PRIMARY KEY,
  clan_id UUID NOT NULL REFERENCES clans(id) ON DELETE CASCADE,
  user_id UUID NOT NULL UNIQUE,
  role VARCHAR(20) NOT NULL DEFAULT 'MEMBER', -- LEADER, MODERATOR, MEMBER
  joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_clan_members_clan_id ON clan_members(clan_id);
CREATE INDEX idx_clan_members_user_id ON clan_members(user_id);