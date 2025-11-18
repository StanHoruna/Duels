ALTER TABLE players
    ALTER COLUMN win_amount TYPE INTEGER
        USING ROUND(win_amount);