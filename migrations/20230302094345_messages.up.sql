CREATE TABLE message
(
    id              SERIAL,
    connection_id   VARCHAR,
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
