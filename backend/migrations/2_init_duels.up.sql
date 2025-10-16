CREATE TABLE IF NOT EXISTS duels
(
    id                     UUID PRIMARY KEY,
    owner_id               UUID        NOT NULL,
    room_number            INTEGER     NULL,
    players_count          INTEGER     NOT NULL DEFAULT 0,
    refunded_players_count INTEGER     NOT NULL DEFAULT 0,
    winners_count          INTEGER     NOT NULL DEFAULT 0,

    username               VARCHAR(17) NOT NULL,

    status                 INTEGER     NOT NULL DEFAULT 0,
    image_url              TEXT        NULL,
    bg_url                 TEXT        NULL,

    question               TEXT        NOT NULL,
    duel_price             INTEGER     NOT NULL,
    commission             INTEGER     NOT NULL,
    duel_info              JSONB       NULL,
    event_date             TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    final_result           INTEGER     NULL,
    cancellation_reason    TEXT        NULL,

    created_at             TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS duels_room_number_uq
    ON duels (room_number) WHERE room_number IS NOT NULL;

CREATE INDEX IF NOT EXISTS duels_owner_id_idx ON duels (owner_id);
CREATE INDEX IF NOT EXISTS duels_status_idx ON duels (status);
CREATE INDEX IF NOT EXISTS duels_event_date_idx ON duels (event_date);
CREATE INDEX IF NOT EXISTS duels_created_at_idx ON duels (created_at);

CREATE TABLE IF NOT EXISTS players
(
    id           UUID PRIMARY KEY,
    user_id      UUID        NOT NULL,
    duel_id      UUID        NOT NULL,
    win_amount   INTEGER     NOT NULL DEFAULT 0,
    answer       INTEGER     NOT NULL,
    final_status SMALLINT    NOT NULL DEFAULT 0,
    is_winner    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT players_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT players_duel_fk FOREIGN KEY (duel_id) REFERENCES duels (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS players_duel_user_uq
    ON players (duel_id, user_id);

CREATE INDEX IF NOT EXISTS players_duel_id_idx ON players (duel_id);
CREATE INDEX IF NOT EXISTS players_user_id_idx ON players (user_id);
CREATE INDEX IF NOT EXISTS players_final_status_idx ON players (final_status);
CREATE INDEX IF NOT EXISTS players_created_at_idx ON players (created_at);

CREATE TABLE IF NOT EXISTS transactions
(
    signature CHAR(88) PRIMARY KEY,
    tx_type   SMALLINT NOT NULL
);

CREATE INDEX IF NOT EXISTS transactions_tx_type_idx
    ON transactions (tx_type);