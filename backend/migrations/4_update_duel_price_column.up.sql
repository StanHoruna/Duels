ALTER TABLE duels
    ALTER COLUMN duel_price TYPE NUMERIC(15, 9)
        USING duel_price::NUMERIC(15, 9);