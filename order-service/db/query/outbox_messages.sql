-- name: CreateOutboxMessage :exec
INSERT INTO outbox_messages (
    id,
    aggregate_id,
    event_type,
    payload,
    created_at,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: GetPendingOutboxMessages :many
SELECT *
FROM outbox_messages
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1;

-- name: MarkOutboxMessageProcessed :exec
UPDATE outbox_messages
SET status = 'processed',
    processed_at = $2
WHERE id = $1;

-- name: MarkOutboxMessageFailed :exec
UPDATE outbox_messages
SET status = 'failed',
    attempt_count = attempt_count + 1,
    error_message = $2
WHERE id = $1;

-- name: GetOutboxMessageByID :one
SELECT *
FROM outbox_messages
WHERE id = $1;

-- name: DeleteOutboxMessage :exec
DELETE FROM outbox_messages
WHERE id = $1;