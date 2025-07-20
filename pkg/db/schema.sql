-- Player represents a registered player.
CREATE TABLE IF NOT EXISTS player (
    id BIGSERIAL PRIMARY KEY,
    -- Name must be a string of english letters or numbers of the length between 2 and 60 symbols.
    name VARCHAR(60) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(60) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- Session represents an active or expired session.
-- The expired session will be deleted after the player tries to login
-- using the Cookie with the id of the expired session.
CREATE TABLE IF NOT EXISTS session (
    id TEXT PRIMARY KEY,
    player_id BIGINT NOT NULL UNIQUE REFERENCES player(id),
    -- The session will be active for a one day.
    expires_at TIMESTAMP NOT NULL DEFAULT (now() + interval '24 hours')
);
