CREATE TABLE payment (
    id UUID PRIMARY KEY NOT NULL,
    amount DECIMAL(19, 4),
    status VARCHAR(255),
    error_message VARCHAR(255),
    event_id VARCHAR(255) NOT NULL,
    external_service_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payment_id ON payment(id);
CREATE INDEX idx_payment_event_id ON payment(event_id);