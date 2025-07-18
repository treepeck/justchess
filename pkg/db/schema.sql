CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Session represents an active or expired session.
-- The expired session will be deleted after the client tries to login
-- using the Cookie with the id of the expired session.
CREATE TABLE IF NOT EXISTS session (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    -- The session will be active for a one day.
    expires_at TIMESTAMP NOT NULL DEFAULT (now() + interval '24 hours')
);