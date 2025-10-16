CREATE TABLE IF NOT EXISTS notifications
(
    id         uuid PRIMARY KEY,
    user_id    uuid        NOT NULL,
    type       integer     NOT NULL,
    data       jsonb       NOT NULL DEFAULT '{}'::jsonb,
    is_read    boolean     NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_notifications_user
        FOREIGN KEY (user_id) REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created_at
    ON notifications (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications (user_id, created_at DESC)
    WHERE is_read = false;

CREATE INDEX IF NOT EXISTS idx_notifications_data_gin
    ON notifications USING GIN (data jsonb_path_ops);