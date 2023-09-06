dsn: "postgres://127.0.0.1/go_restful?sslmode=disable&user=postgres&password=postgres" # Postgres dsn.
self_apps:
  - self_app_id: "" # Your app self identifier.
    self_device_secret: "" # Secret key for the device created at the developer portal.
    self_storage_dir: "/data" # The storage folder you want to use for Self sessions.
    self_storage_key: "" # Key to be used for encrypting local storage.
    self_env: "sandbox" # Self environment you want to point to.
    message_notification_url: "" # The URL to send incoming messages notifications to.
jwt_signing_key: "" # Signing key for JWT.
jwt_expiration: 24 # JWT expiration in hours.
refresh_token_expiration: 128 # JWT refresh expiration in hours.
user: "demo" # User to be used for JWT authentication.
password: "pass" # Password to be used for JWT authentication.
serve_docs: "true" # Serves the docs from the local server on localhost:8080/docs
