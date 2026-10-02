CREATE TABLE vps_hosting (
    id UUID PRIMARY KEY NOT NULL,
    hosting_name VARCHAR(255),
    api_token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vps_hosting_id ON vps_hosting(id);
CREATE INDEX idx_vps_hosting_hosting_name ON vps_hosting(hosting_name);