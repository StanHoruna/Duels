DROP INDEX IF EXISTS duels_created_at_idx;
DROP INDEX IF EXISTS duels_event_date_idx;
DROP INDEX IF EXISTS duels_status_idx;
DROP INDEX IF EXISTS duels_owner_id_idx;
DROP INDEX IF EXISTS duels_room_number_uq;

DROP TABLE IF EXISTS duels;

DROP INDEX IF EXISTS players_created_at_idx;
DROP INDEX IF EXISTS players_final_status_idx;
DROP INDEX IF EXISTS players_user_id_idx;
DROP INDEX IF EXISTS players_duel_id_idx;
DROP INDEX IF EXISTS players_duel_user_uq;

ALTER TABLE IF EXISTS players
    DROP CONSTRAINT IF EXISTS players_duel_fk;
ALTER TABLE IF EXISTS players
    DROP CONSTRAINT IF EXISTS players_user_fk;

DROP TABLE IF EXISTS players;

DROP INDEX IF EXISTS transactions_tx_type_idx;
DROP TABLE IF EXISTS transactions;