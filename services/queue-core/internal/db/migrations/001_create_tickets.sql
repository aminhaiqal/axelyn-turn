CREATE TABLE tickets (
    id BIGSERIAL PRIMARY KEY,
    queue_id BIGINT NOT NULL,
    customer_name TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'waiting',
    priority INT DEFAULT 1,
    estimated_time INT DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tickets_queue_status ON tickets(queue_id, status);
