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