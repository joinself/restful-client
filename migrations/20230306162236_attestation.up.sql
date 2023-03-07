CREATE TABLE attestation
(
    id              VARCHAR NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    fact_id         VARCHAR,
    body            TEXT,
    value           TEXT,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    CONSTRAINT fk_fact
      FOREIGN KEY(fact_id) 
	  REFERENCES fact(id)
);
