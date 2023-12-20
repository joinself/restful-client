CREATE TABLE connection
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    selfid     VARCHAR NOT NULL,
    appid      VARCHAR NOT NULL,
    name       VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(selfid, appid)
);
