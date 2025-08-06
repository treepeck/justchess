-- Player represents a registered player.
CREATE TABLE IF NOT EXISTS player (
    id CHAR(26) PRIMARY KEY,
    -- Name must be a string of english letters or numbers of the length
    -- between 2 and 60 symbols.
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
    id CHAR(26) PRIMARY KEY,
    player_id CHAR(26) NOT NULL UNIQUE REFERENCES player(id),
    -- The session will be active for a one day.
    expires_at TIMESTAMP NOT NULL DEFAULT (now() + interval '24 hours')
);

DO $$ BEGIN
    CREATE DOMAIN GAME_RESULT AS INT
    CHECK (VALUE IN (0, 1, 2, 3, 4, 5, 6, 7, 8));
    EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Game represents an active or already overed game.
CREATE TABLE IF NOT EXISTS game (
    id CHAR(26) PRIMARY KEY,
    white_id CHAR(26) NOT NULL UNIQUE REFERENCES player(id),
    black_id CHAR(26) NOT NULL UNIQUE REFERENCES player(id),
    result GAME_RESULT NOT NULL
);