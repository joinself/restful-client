CREATE TABLE app
(
    id                  VARCHAR PRIMARY KEY,
    device_secret       VARCHAR NOT NULL,
    name                VARCHAR NOT NULL,
    env                 VARCHAR NOT NULL,
    callback            VARCHAR NOT NULL,
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    UNIQUE(name)
);
