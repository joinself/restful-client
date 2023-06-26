CREATE TABLE request
(
    id              VARCHAR NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    connection_id   INTEGER NOT NULL,
    type            VARCHAR(255) NOT NULL,
    status          VARCHAR(255) DEFAULT 'requested',
    facts           TEXT NOT NULL,
    auth            BOOLEAN,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    CONSTRAINT fk_connection
      FOREIGN KEY(connection_id)
	  REFERENCES connection(id)
);
