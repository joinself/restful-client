ALTER TABLE message DROP CONSTRAINT fk_message_rooms;
ALTER TABLE message DROP COLUMN gid; 

DROP TABLE room_connection;
DROP TABLE room;
