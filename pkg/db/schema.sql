-- Represents a registered human player (or engine).
CREATE TABLE IF NOT EXISTS player (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	mail TEXT NOT NULL UNIQUE, 
	name TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	-- Whether this player is a chess engine.
	is_engine BOOLEAN NOT NULL DEFAULT false,
	created_at TIMESTAMP NOT NULL DEFAULT now(),
	updated_at TIMESTAMP NOT NULL DEFAULT now()
);

INSERT INTO player (id, mail, name, password_hash, is_engine)
VALUES ('ccaf962b-855e-49da-b85f-7e8bba0edae2', '', 'Stockfish', '', true)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS tokenregistration (
	token TEXT PRIMARY KEY,
	mail TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tokenreset (
	token TEXT PRIMARY KEY,
	player_id UUID NOT NULL REFERENCES player(id),
	password_hash TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- 0 - White;
-- 1 - Black;
-- 2 - None.
DO $$ BEGIN
	CREATE DOMAIN COLOR AS SMALLINT
	CHECK (VALUE IN (0, 1, 2));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- 0 - Unknown;
-- 1 - Checkmate;
-- 2 - Timeout;
-- 3 - Stalemate;
-- 4 - InsufficientMaterial;
-- 5 - FiftyMoves;
-- 6 - Repetition;
-- 7 - Resignation;
-- 8 - Agreement.  
DO $$ BEGIN
	CREATE DOMAIN GAME_RESULT AS SMALLINT
	CHECK (VALUE IN (0, 1, 2, 3, 4, 5, 6, 7, 8));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- 0 - Basic;
-- 1 - Engine.
DO $$ BEGIN 
	CREATE DOMAIN GAME_MODE AS SMALLINT
	CHECK (VALUE IN (0, 1));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Represents completed game.
CREATE TABLE IF NOT EXISTS game (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	white_id UUID NOT NULL REFERENCES player(id),
	black_id UUID NOT NULL REFERENCES player(id), 
	time_control SMALLINT NOT NULL,
	time_bonus SMALLINT NOT NULL,
	result GAME_RESULT NOT NULL,
	winner COLOR NOT NULL,
	-- Each moves take 32 bits:
	--   0-15: Move (see movegen.go);
	--   16-31: Remaining time on a player's clock in seconds.
	moves INTEGER[] NOT NULL,
	mode GAME_MODE NOT NULL
);