CREATE TABLE payment (
    id UUID PRIMARY KEY NOT NULL,
    amount DECIMAL(19, 4),
    status VARCHAR(255),
    error_message VARCHAR(255),
    event_id VARCHAR(255) NOT NULL,
    external_service_id BIGINT,
    vps_hosting_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_payment_vps_hosting FOREIGN KEY (vps_hosting_id) REFERENCES vps_hosting(id)
);
CREATE INDEX idx_payment_id ON payment(id);
CREATE INDEX idx_payment_event_id ON payment(event_id);
CREATE INDEX idx_payment_vps_hosting_id ON payment(vps_hosting_id);


CREATE TABLE refund (
    id UUID PRIMARY KEY NOT NULL,
    amount DECIMAL(19, 4),
    reason VARCHAR(255),
    status VARCHAR(255),
    payment_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_refund_id ON refund(id);
CREATE INDEX idx_refund_payment_id ON refund(payment_id);


CREATE TABLE vps_hosting (
    id UUID PRIMARY KEY NOT NULL,
    hosting_name VARCHAR(255),
    api_token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_vps_hosting_id ON vps_hosting(id);
CREATE INDEX idx_vps_hosting_hosting_name ON vps_hosting(hosting_name);