CREATE TABLE request (
    id TEXT NOT NULL UNIQUE DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-a' || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
    connection_id INTEGER,
    type TEXT NOT NULL,
    status TEXT DEFAULT 'requested',
    facts TEXT, --NOT NULL,
    auth INTEGER,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY(connection_id) REFERENCES connection(id)
);
