CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    public_address CHAR(44) UNIQUE,
    image_url VARCHAR(100),
    created_at timestamp without time zone NOT NULL DEFAULT current_timestamp,
    updated_at timestamp without time zone NOT NULL DEFAULT current_timestamp
);
