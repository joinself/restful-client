CREATE TABLE fact
(
    id TEXT NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-a' || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    connection_id INTEGER NOT NULL,
    request_id    TEXT,
    iss           TEXT NOT NULL DEFAULT '',
    cid           TEXT NOT NULL DEFAULT '',
    jti           TEXT NOT NULL DEFAULT '',
    status        TEXT DEFAULT 'requested',
    source        TEXT NOT NULL DEFAULT '*',
    fact          TEXT NOT NULL,
    body          TEXT,
    iat           DATETIME NOT NULL,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    FOREIGN KEY(connection_id) REFERENCES connection(id),
    FOREIGN KEY(request_id) REFERENCES request(id)
);
