CREATE TABLE metric
(
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    appid               VARCHAR NOT NULL,
    uuid                INTEGER NOT NULL,
    recipient           VARCHAR NOT NULL,
    actions             VARCHAR NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    UNIQUE(uuid)
);
