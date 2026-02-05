CREATE TABLE IF NOT EXISTS posts(
    id UUID PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    body VARCHAR(2000) NOT NULL,
    comments_enabled BOOLEAN NOT NULL DEFAULT true,
    author_id UUID NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
