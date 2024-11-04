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

-- custom types creation.
-- 0 - Bullet 1min;
-- 1 - Blitz 3min;
-- 2 - Rapid 10min.
DO $$ BEGIN
	CREATE DOMAIN TIME_CONTROL AS INTEGER
  CHECK (VALUE IN (0, 1, 2));
EXCEPTION
	WHEN duplicate_object THEN 
	RAISE NOTICE 'time_control already created, skipping innitialization';
END $$;

DO $$ BEGIN
	 -- time bonus represented in seconds, 0 - no time bonus.
    CREATE DOMAIN TIME_BONUS AS INTEGER
    CHECK (VALUE IN (0, 1, 2, 10));
EXCEPTION
	WHEN duplicate_object THEN 
	RAISE NOTICE 'time_bonus already created, skipping innitialization';
END $$;

-- 0 - Checkmate;
-- 1 - Resignation;
-- 2 - Timeout;
-- 3 - Stalemate;
-- 4 - InsufficientMaterial;
-- 5 - FiftyMoves;
-- 6 - Repetition;
-- 7 - Agreement.  
DO $$ BEGIN
	 -- game result describes the ways chess game can end. 
    CREATE DOMAIN GAME_RESULT AS INTEGER
    CHECK (VALUE IN (0, 1, 2, 3, 4, 5, 6, 7));
EXCEPTION
	WHEN duplicate_object THEN 
	RAISE NOTICE 'game_result already created, skipping innitialization';
END $$;

CREATE TABLE IF NOT EXISTS games(
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  black_id VARCHAR(36),
  white_id VARCHAR(36),
  control TIME_CONTROL NOT NULL,
  bonus TIME_BONUS NOT NULL,
  result GAME_RESULT NOT NULL,
  moves JSONB,
  played_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (black_id) REFERENCES users(id),
  FOREIGN KEY (white_id) REFERENCES users(id)
);