CREATE TABLE admins (
    id         VARCHAR(26) PRIMARY KEY,
    username   VARCHAR(50) UNIQUE NOT NULL,
    password   TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);