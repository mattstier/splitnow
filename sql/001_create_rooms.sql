CREATE TABLE rooms(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    -- user_id of the creator, not a foreign key (denormalized) 
    -- because users table is going to be in the auth service 
    created_by INT NOT NULL, 
    created_at TIMESTAMPTZ DEFAULT NOW()
);
