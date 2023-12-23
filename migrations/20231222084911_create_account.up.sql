CREATE TABLE account
(
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    user_name           VARCHAR NOT NULL,
    hashed_password     VARCHAR NOT NULL,
    salt                VARCHAR NOT NULL,
    resources           TEXT NOT NULL DEFAULT '{}',
    created_at          TIMESTAMP NOT NULL,
    updated_at          TIMESTAMP NOT NULL,
    UNIQUE(user_name)
);
