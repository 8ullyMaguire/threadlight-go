BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(30) UNIQUE NOT NULL,
    display_name    VARCHAR(100),
    bio             TEXT,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    trust_level     SMALLINT DEFAULT 0 NOT NULL,
    trust_score     REAL DEFAULT 0 NOT NULL,
    reputation      BIGINT DEFAULT 0 NOT NULL,
    invited_by      BIGINT REFERENCES users(id),
    invite_code     VARCHAR(32) UNIQUE,
    credits         BIGINT DEFAULT 0 NOT NULL,
    is_active       BOOLEAN DEFAULT TRUE,
    last_active_at  TIMESTAMPTZ,
    public_key      TEXT,
    actor_id        VARCHAR(255) UNIQUE,
    is_local        BOOLEAN DEFAULT TRUE,
    onboarding_stage SMALLINT DEFAULT 0,
    proximity_opt_out BOOLEAN DEFAULT FALSE,
    location_hash   VARCHAR(64),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_invite_code ON users(invite_code);

COMMIT;
