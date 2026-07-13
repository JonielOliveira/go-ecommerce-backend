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

CREATE TYPE order_status AS ENUM (
    'PENDING',
    'PAID',
    'CANCELED'
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    customer_id UUID NOT NULL,

    status order_status NOT NULL DEFAULT 'PENDING',
    total_amount NUMERIC(12, 2) NOT NULL,

    paid_at TIMESTAMPTZ NULL,
    canceled_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_orders_customer
        FOREIGN KEY (customer_id)
        REFERENCES users (id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_orders_total_amount
        CHECK (total_amount >= 0),

    CONSTRAINT chk_orders_status_timestamps
        CHECK (
            (
                status = 'PENDING'
                AND paid_at IS NULL
                AND canceled_at IS NULL
            )
            OR
            (
                status = 'PAID'
                AND paid_at IS NOT NULL
                AND canceled_at IS NULL
            )
            OR
            (
                status = 'CANCELED'
                AND paid_at IS NULL
                AND canceled_at IS NOT NULL
            )
        )
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    order_id UUID NOT NULL,
    product_id UUID NOT NULL,

    quantity INTEGER NOT NULL,
    unit_price NUMERIC(10, 2) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id)
        REFERENCES orders (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_order_items_product
        FOREIGN KEY (product_id)
        REFERENCES products (id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_order_items_quantity
        CHECK (quantity > 0),

    CONSTRAINT chk_order_items_unit_price
        CHECK (unit_price >= 0),

    CONSTRAINT uq_order_items_order_product
        UNIQUE (order_id, product_id)
);

CREATE INDEX idx_orders_customer_id
    ON orders (customer_id);

CREATE INDEX idx_orders_status
    ON orders (status);

CREATE INDEX idx_orders_created_at
    ON orders (created_at DESC, id DESC);

CREATE INDEX idx_orders_customer_created_at
    ON orders (customer_id, created_at DESC, id DESC);

CREATE INDEX idx_order_items_order_id
    ON order_items (order_id);

CREATE INDEX idx_order_items_product_id
    ON order_items (product_id);
    