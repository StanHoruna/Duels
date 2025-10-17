ALTER TABLE duels
    ALTER COLUMN duel_price TYPE INTEGER
        USING ROUND(duel_price);