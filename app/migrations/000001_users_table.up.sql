CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    username VARCHAR(15)
);

INSERT IGNORE INTO users(id, username) VALUES
(1, 'ASDASd'),
(2, 'asd'),
(3, 'wevbwb'),