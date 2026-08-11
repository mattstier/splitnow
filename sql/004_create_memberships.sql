CREATE TABLE memberships (
    room_id INT REFERENCES rooms(id) ON DELETE CASCADE,
    -- user_id from auth service, not a foreign key (users live in a different DB)
    user_id INT NOT NULL,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (room_id, user_id)
);

-- index to speed up querying the rooms of a given user and vice versa 
CREATE INDEX idx_memberships_user ON memberships (user_id);
