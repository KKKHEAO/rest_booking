CREATE TABLE tables (
    id     UUID PRIMARY KEY,
    number INT  NOT NULL UNIQUE,
    seats  INT  NOT NULL CHECK (seats > 0)
);

CREATE TABLE bookings (
    id         UUID        PRIMARY KEY,
    table_id   UUID        NOT NULL REFERENCES tables (id),
    guest_name TEXT        NOT NULL,
    guests     INT         NOT NULL CHECK (guests > 0),
    starts_at  TIMESTAMPTZ NOT NULL,
    ends_at    TIMESTAMPTZ NOT NULL,
    status     TEXT        NOT NULL DEFAULT 'confirmed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (starts_at < ends_at)
);

CREATE INDEX idx_bookings_table_id ON bookings (table_id);
