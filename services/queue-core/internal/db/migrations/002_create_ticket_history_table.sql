-- 02_create_ticket_history_table.sql

CREATE TABLE ticket_history (
    id BIGINT PRIMARY KEY,
    queue_id BIGINT NOT NULL,
    customer_name TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    priority INT,
    estimated_time INT,
    version BIGINT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    archived_at TIMESTAMPTZ DEFAULT NOW()
);

-- History table index for reporting
CREATE INDEX idx_ticket_history_queue_id ON ticket_history(queue_id);
