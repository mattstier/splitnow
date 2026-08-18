CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    room_id INT REFERENCES rooms(id), 
    sender TEXT NOT NULL,
    content TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT FALSE, 
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- index to speed up querying old messages of a room
CREATE INDEX idx_messages_room_created ON messages (room_id, created_at);
