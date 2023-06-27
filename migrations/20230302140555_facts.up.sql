CREATE TABLE fact
(
    id              VARCHAR NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    connection_id   INTEGER NOT NULL,
    request_id      VARCHAR(255) NOT NULL,
    iss             VARCHAR(255) DEFAULT '' NOT NULL,
    cid             VARCHAR(255) DEFAULT '' NOT NULL,
    jti             VARCHAR(255) DEFAULT '' NOT NULL,
    status          VARCHAR(255) DEFAULT 'requested',
    source          VARCHAR(255) DEFAULT '*' NOT NULL,
    fact            VARCHAR(255) NOT NULL,
    body            TEXT,
    iat             TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    CONSTRAINT fk_connection
      FOREIGN KEY(connection_id) 
	  REFERENCES connection(id),
    CONSTRAINT fk_request
      FOREIGN KEY(request_id)
	  REFERENCES request(id)
);
