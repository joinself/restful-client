CREATE TABLE fact
(
    id              VARCHAR PRIMARY KEY,
    connection_id   VARCHAR,
    iss             VARCHAR(255) DEFAULT '' NOT NULL,
    source          VARCHAR(255) NOT NULL,
    body            TEXT,
    iat             TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    CONSTRAINT fk_connection
      FOREIGN KEY(connection_id) 
	  REFERENCES connection(id)
);
