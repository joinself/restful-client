CREATE TABLE apikey
(
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    appid               VARCHAR NOT NULL,
    token               VARCHAR NOT NULL,
    name                VARCHAR NOT NULL,
    scope               VARCHAR NOT NULL,
    deleted             INTEGER DEFAULT 0 NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP NOT NULL,
    UNIQUE(name)
);
