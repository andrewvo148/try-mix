CREATE TABLE outbox_messages (
    id UUID PRIMARY KEY,
    aggregate_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    error_message TEXT NULL
);


-- Indexes for better performance
-- Create indexes
CREATE INDEX idx_outbox_messages_status ON outbox_messages(status);
CREATE INDEX idx_outbox_messages_created_at ON outbox_messages(created_at);
CREATE INDEX idx_outbox_messages_aggregate_id ON outbox_messages(aggregate_id);