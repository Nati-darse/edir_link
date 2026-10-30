CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    full_name VARCHAR(150) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(255),

    password_hash TEXT NOT NULL,

    role VARCHAR(30) NOT NULL
        CHECK (role IN ('SUPER_ADMIN', 'ORGANIZER', 'MEMBER')),

    must_change_password BOOLEAN NOT NULL DEFAULT FALSE,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX users_email_unique
ON users (LOWER(email))
WHERE email IS NOT NULL;

CREATE UNIQUE INDEX users_phone_unique
ON users (phone)
WHERE phone IS NOT NULL;


CREATE TABLE edirs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(150) NOT NULL,

    organizer_id UUID NOT NULL
        REFERENCES users(id),

    monthly_contribution_amount NUMERIC(12,2) NOT NULL
        CHECK (monthly_contribution_amount >= 0),

    due_day SMALLINT NOT NULL
        CHECK (due_day BETWEEN 1 AND 28),

    support_amount NUMERIC(12,2) NOT NULL DEFAULT 0
        CHECK (support_amount >= 0),

    penalty_percentage NUMERIC(5,2) NOT NULL DEFAULT 0
        CHECK (penalty_percentage >= 0),

    status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'ACTIVE', 'REJECTED')),

    rejection_reason TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    edir_id UUID NOT NULL
        REFERENCES edirs(id)
        ON DELETE CASCADE,

    member_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    gender VARCHAR(20)
        CHECK (gender IN ('MALE', 'FEMALE', 'OTHER')),

    address TEXT,

    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    UNIQUE (edir_id, member_id)
);

CREATE INDEX idx_edirs_organizer
ON edirs (organizer_id);

CREATE INDEX idx_memberships_edir
ON memberships (edir_id);

CREATE INDEX idx_memberships_member
ON memberships (member_id);