#!/bin/bash -e

#exec > >(tee -a /var/log/app/entry.log|logger -t server -s 2>/dev/console) 2>&1

echo "[`date`] Running entrypoint script..."

if [[ -z ${CONFIG_FILE} ]]; then
  export CONFIG_FILE=./config/local.yml
fi
echo "[`date`] Loading configuration from ${CONFIG_FILE}..."

if [[ -z ${APP_DSN} ]]; then
  export APP_DSN=`sed -n 's/^dsn:[[:space:]]*"\(.*\)"/\1/p' ${CONFIG_FILE}`
fi

echo "[`date`] Running database migrations..."
until migrate -database "${APP_DSN}" -path ./migrations up; do
  echo "[`date`] Waiting for database..."
  sleep 5
done

echo "[`date`] Starting server..."
./server -config ${CONFIG_FILE}
