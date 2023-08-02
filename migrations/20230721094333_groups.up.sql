ALTER TABLE message ALTER COLUMN connection_id DROP NOT NULL;
ALTER TABLE message ADD COLUMN gid VARCHAR(255) UNIQUE;

CREATE TABLE room
(
    id              SERIAL PRIMARY KEY,
    appid           VARCHAR NOT NULL,
    gid             VARCHAR(255) UNIQUE NOT NULL,
    name            VARCHAR(255) DEFAULT '' NOT NULL,
    status          VARCHAR(255) DEFAULT '' NOT NULL,
    icon_link       TEXT DEFAULT '' NOT NULL,
    icon_mime       VARCHAR(255) DEFAULT '',
    icon_key        TEXT,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL
);

CREATE TABLE room_connection
(
    id              SERIAL PRIMARY KEY,
    room_id         INTEGER NOT NULL,
    connection_id   INTEGER NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    CONSTRAINT fk_connection
      FOREIGN KEY(connection_id) 
	  REFERENCES connection(id),
    CONSTRAINT fk_room
      FOREIGN KEY(room_id) 
	  REFERENCES room(id)
);

ALTER TABLE message ADD 
    CONSTRAINT fk_message_rooms 
    FOREIGN KEY (gid) 
    REFERENCES room (gid)
    ON DELETE SET NULL;
