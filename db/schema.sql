CREATE TABLE IF NOT EXISTS users(
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  name VARCHAR(36) NOT NULL,
  password VARCHAR(36) NOT NULL,
  blitz_rating INT DEFAULT 400,
  rapid_rating INT DEFAULT 400,
  bullet_rating INT DEFAULT 400,
  games_count INT DEFAULT 0,
  likes INT DEFAULT 0,
  is_deleted BOOLEAN DEFAULT FALSE,
  registered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_visit TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- create custom types
DO $$ BEGIN
	CREATE TYPE TIME_CONTROL AS ENUM ('blitz', 'bullet', 'rapid');
EXCEPTION
	WHEN duplicate_object THEN 
	RAISE NOTICE 'time_control already created, skipping innitialization';
END $$;

DO $$ BEGIN
	CREATE TYPE GAME_STATUS AS ENUM (
    'white_won', 'black_won',
    'draw', 'continues', 'waiting'
  );
EXCEPTION
	WHEN duplicate_object THEN 
	RAISE NOTICE 'game_status already created, skipping innitialization';
END $$;

DO $$ BEGIN
	 -- time bonus represented in seconds, 0 - no time bonus
    CREATE DOMAIN TIME_BONUS AS INTEGER
    CHECK (VALUE IN (0, 1, 2, 10));
EXCEPTION
	WHEN duplicate_object THEN 
	RAISE NOTICE 'game_status already created, skipping innitialization';
END $$;

CREATE TABLE IF NOT EXISTS games(
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  black_id VARCHAR(36),
  white_id VARCHAR(36),
  control TIME_CONTROL NOT NULL,
  bonus TIME_BONUS NOT NULL,
  status GAME_STATUS NOT NULL DEFAULT 'waiting',
  moves JSONB,
  played_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (black_id) REFERENCES users(id),
  FOREIGN KEY (white_id) REFERENCES users(id)
);