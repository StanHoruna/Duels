ALTER TABLE players
    ALTER COLUMN win_amount TYPE NUMERIC(15, 9)
        USING win_amount::NUMERIC(15, 9);