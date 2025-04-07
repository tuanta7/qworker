CREATE TABLE IF NOT EXISTS private.user
(
    id             UUID PRIMARY KEY,
    username       VARCHAR(255)  NOT NULL UNIQUE,
    full_name      VARCHAR(1000),
    phone_number   VARCHAR(20),
    email          VARCHAR(1000) NOT NULL,
    email_verified BOOLEAN                DEFAULT false,
    active         BOOLEAN                DEFAULT false,
    data           TEXT,
    source_id      SERIAL,
    created_at     TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP     NOT NULL DEFAULT NOW(),
    FOREIGN KEY (source_id) REFERENCES private.connector (id)
);

CREATE TABLE private.mapper
(
    external_id VARCHAR(255),
    full_name   VARCHAR(255),
    updated_at  VARCHAR(255)
);