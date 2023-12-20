CREATE TABLE message
(
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    connection_id   INTEGER NOT NULL,
    iss             VARCHAR(255) DEFAULT '' NOT NULL,
    cid             VARCHAR(255) DEFAULT '' NOT NULL,
    rid             VARCHAR(255) DEFAULT '',
    body            TEXT,
    iat             TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    CONSTRAINT fk_connection
      FOREIGN KEY(connection_id) 
	  REFERENCES connection(id)
);
