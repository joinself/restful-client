CREATE TABLE signature
(
    id              TEXT NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-a' || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    app_id           VARCHAR NOT NULL,
    self_id          VARCHAR NOT NULL,
    description     VARCHAR NOT NULL,
    status          VARCHAR NOT NULL,
    data            TEXT,
    signature       VARCHAR,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL
);
