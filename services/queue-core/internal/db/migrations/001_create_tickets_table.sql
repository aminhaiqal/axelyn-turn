-- 01_create_tickets_table.sql

CREATE TABLE tickets (
    id BIGSERIAL PRIMARY KEY,
    queue_id BIGINT NOT NULL,
    customer_name TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'waiting',
    priority INT DEFAULT 1,
    estimated_time INT DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Composite index for general filtering
CREATE INDEX idx_tickets_queue_status ON tickets(queue_id, status);

-- Partial indexes for active queues only (faster, smaller)
CREATE INDEX idx_tickets_waiting 
    ON tickets(queue_id) 
    WHERE status = 'waiting';

CREATE INDEX idx_tickets_in_progress 
    ON tickets(queue_id) 
    WHERE status = 'in_progress';
