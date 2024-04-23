CREATE TABLE call
(
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    appid               VARCHAR NOT NULL,
    selfid              VARCHAR NOT NULL,
    peer_info           VARCHAR NOT NULL,
    call_id             VARCHAR NOT NULL,
    status              VARCHAR NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    deleted_at          TIMESTAMP NOT NULL,
    UNIQUE(name)
);
