#!/bin/bash

echo "[`date`] Running database migrations..."
until migrate -path=/srv/migrations/ -database sqlite3://${RESTFUL_CLIENT_STORAGE_DIR}/client.db up; do
  echo "[`date`] Waiting for database..."
  sleep 5
done

echo "[`date`] Starting server..."
/srv/restful-client
