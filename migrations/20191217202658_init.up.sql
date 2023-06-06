CREATE TABLE connection
(
    id         SERIAL
    selfid     VARCHAR NOT NULL,
    appid      VARCHAR NOT NULL,
    name       VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
    PRIMARY KEY(id, appid)
);
