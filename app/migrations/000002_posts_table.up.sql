CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    body VARCHAR(2000) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);