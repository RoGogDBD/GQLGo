CREATE TABLE IF NOT EXISTS comments(
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id),
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    body VARCHAR(2000) NOT NULL,
    depth INTEGER NOT NULL DEFAULT 0,
    children_count INTEGER NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS comments_post_parent_created_idx ON comments(post_id, parent_id, created_at, id);
