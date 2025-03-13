CREATE TABLE private.user
(
    id         UUID PRIMARY KEY,
    data       TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE private.mapping
(
);