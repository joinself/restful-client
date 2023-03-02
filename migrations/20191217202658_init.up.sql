CREATE TABLE connection
(
    id         VARCHAR PRIMARY KEY,
    name       VARCHAR NOT NULL,
    selfid     VARCHAR UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
