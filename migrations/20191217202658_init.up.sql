CREATE TABLE connection
(
    id         SERIAL PRIMARY KEY,
    selfid     VARCHAR NOT NULL,
    appid      VARCHAR NOT NULL,
    name       VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(selfid, appid)
);
