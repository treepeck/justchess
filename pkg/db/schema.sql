-- Stores players with verified by mail accounts.
CREATE TABLE IF NOT EXISTS player (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	mail TEXT NOT NULL UNIQUE, 
	name TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT now(),
	updated_at TIMESTAMP NOT NULL DEFAULT now()
);

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

