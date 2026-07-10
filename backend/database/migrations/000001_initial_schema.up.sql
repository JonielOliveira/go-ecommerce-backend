CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE user_role AS ENUM (
    'customer',
    'admin'
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    price NUMERIC(10, 2) NOT NULL,
    stock INTEGER NOT NULL,
    category_id UUID NULL,
    image_url TEXT NULL,

    active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    name VARCHAR(255) NOT NULL,
    email CITEXT NOT NULL UNIQUE,
    email_verified_at TIMESTAMPTZ NULL,
    avatar_url TEXT NULL,

    role user_role NOT NULL DEFAULT 'customer',
    active BOOLEAN NOT NULL DEFAULT TRUE,

    last_login_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE user_password_credentials (
    user_id UUID PRIMARY KEY,

    password_hash TEXT NOT NULL,
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_password_credentials_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);
