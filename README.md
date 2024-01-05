# Go RESTful Joinself Client

This joinself client is designed to act as a proxy between your app and The Self Network. It exposes a RESTful API to interact with the basic features Self has to offer.

> :warning: **Please note this software is currently under active development and is subject to backwards incompatible changes at any time.**

## Getting Started

There are a few different ways to run this service depending on what it is you want to do. The easiest method is to use Docker Compose with the example configuration from the repository. If you're looking to contribute back to the project then the `Build from source` section is for you.

### Docker Compose

An example `docker-compose.yml` can be found at the root of this project.

Add your application ID, secret key and any other changes you require then start the service.

```bash
docker compose up
```

### Docker

To run the service directly via Docker amend the environment variables as required and run:

```bash
docker run -it \
  -e RESTFUL_CLIENT_JWT_SIGNING_KEY=secret \
  -e RESTFUL_CLIENT_USER=self \
  -e RESTFUL_CLIENT_PASSWORD=secret \
  -e RESTFUL_CLIENT_STORAGE_DIR=/data \
  -e RESTFUL_CLIENT_STORAGE_KEY=secret \
  -e RESTFUL_CLIENT_APP_ID=<SELF_APP_ID> \
  -e RESTFUL_CLIENT_APP_SECRET=<SELF_APP_SECRET> \
  -e RESTFUL_CLIENT_APP_ENV=sandbox \
  -p 8080:8080 \
  -v restful-client:/data \
  ghcr.io/joinself/restful-client:latest
```

> Note: Replace `SELF_APP_ID` and `SELF_APP_SECRET` with your application ID and secret key.

### Build from source

#### Requirements

- Go 1.21+
- Self OMEMO
- Golang Migrate CLI

##### Self OMEMO

End-to-end encryption protocol. Refer to the [Installing a released version](https://github.com/joinself/self-omemo?tab=readme-ov-file#installing-a-released-version) for installation instructions on different OSS.

##### Database Migration CLI

Database migration tool. https://github.com/golang-migrate/migrate

```bash
  curl -Lo /tmp/migrate.tar.gz https://download.joinself.com/golang-migrate/migrate-sqlite3-4.16.2.tar.gz && \
  tar -zxf /tmp/migrate.tar.gz -C /usr/local/bin
```

#### Build

``` bash
git clone https://github.com/joinself/restful-client.git
cd restful-client

# Replace <SELF_APP_ID> and <SELF_APP_SECRET> with your application ID and secret key.
export RESTFUL_CLIENT_JWT_SIGNING_KEY=secret
export RESTFUL_CLIENT_USER=self
export RESTFUL_CLIENT_PASSWORD=secret
export RESTFUL_CLIENT_STORAGE_DIR=/data
export RESTFUL_CLIENT_STORAGE_KEY=secret
export RESTFUL_CLIENT_APP_ID=<SELF_APP_ID>
export RESTFUL_CLIENT_APP_SECRET=<SELF_APP_SECRET>
export RESTFUL_CLIENT_APP_ENV=sandbox

make migrate
go run cmd/server/main.go
```

The service should now be accessible at `https://localhost:8080`.

## Endpoints

Restful client provides a subset of features to interact with self network, for more details you have the *openapi specification* on `/docs/` folder.

- [x] `GET /healthcheck`: a healthcheck service provided for health checking purpose (needed when implementing a server cluster)
- [x] `POST /v1/login`: authenticates a user and generates a JWT

- [x] `GET /v1/apps`: gets the list of configured apps on the current service

- [x] `GET /v1/apps/:app_id/connections`: returns a paginated list of the existing connections
- [x] `GET /v1/apps/:app_id/connections/:id`: returns the detailed information of an connection
- [x] `POST /v1/apps/:app_id/connections`: creates a new connection
- [x] `PUT /v1/apps/:app_id/connections/:id`: updates an existing connection
- [x] `DELETE /v1/apps/:app_id/connections/:id`: deletes an connection

- [x] `POST /v1/apps/:app_id/connections/:id/messages`: sends a message to the specified connection
- [x] `GET /v1/apps/:app_id/connections/:id/messages`: retrieves the full conversation with a specific connection
- [ ] `GET /v1/apps/:app_id/connections/:id/messages/:id`
- [ ] `PUT /v1/apps/:app_id/connections/:id/messages/:id` updates an existing message
- [ ] `DELETE /v1/apps/:app_id/connections/:id/messages/:id` deletes an existing message from the conversation

- [ ] `GET  /v1/apps/:app_id/groups`
- [ ] `GET  /v1/apps/:app_id/groups/:id`
- [ ] `GET  /v1/apps/:app_id/groups/:id/messages`
- [ ] `GET  /v1/apps/:app_id/groups/:id/messages/:id`
- [ ] `PUT  /v1/apps/:app_id/groups/:id/messages/:id`
- [ ] `GET  /v1/apps/:app_id/groups/:id/messages/:id/responses`

- [x] `POST /v1/apps/:app_id/connections/:cid/requests` : sends a request to the given connection.
- [x] `GET /v1/apps/:app_id/connections/:cid/requests/:id` : retrieves the information about a specific connection, used to retrieve the request status.

- [ ] `POST /v1/apps/:app_id/connections/:id/facts`: issues a fact to a specific connection.
- [x] `GET  /v1/apps/:app_id/connections/:id/facts/:fact_id`: gets the details of an already requested fact.
- [ ] `DELETE /v1/apps/:app_id/connections/:id/facts/:fact_id` deletes a specific fact for a connection


Try the URL `http://localhost:8080/healthcheck` in a browser, and you should see something like `"OK v1.0.0"` displayed.

If you have `cURL` or some API client tools (e.g. [Postman](https://www.getpostman.com/)), you may try the following 
more complex scenarios:

```shell
# authenticate the user via: POST /v1/login
curl -X POST -H "Content-Type: application/json" -d '{"username": "demo", "password": "pass"}' http://localhost:8080/v1/login
# should return a JWT token like: {"token":"...JWT token here..."}

# with the above JWT token, access the connection resources, such as: GET /v1/connections
curl -X GET -H "Authorization: Bearer ...JWT token here..." http://localhost:8080/v1/connections
# should return a list of connection records in the JSON format
```

## Project Layout

The Joinself restful client uses the following project layout:
 
```
.
├── cmd                  main applications of the project
│   └── server           the API server application
├── config               sample configuration files
├── internal             private application and library code
│   ├── connection       connection-related features
│   ├── message          message-related features
│   ├── fact             fact-related features
│   ├── self             self runners to listen for self network events
│   ├── auth             authentication feature
│   ├── config           configuration library
│   ├── entity           entity definitions and domain logic
│   ├── errors           error types and handling
│   ├── healthcheck      healthcheck feature
│   └── test             helpers for testing purpose
├── migrations           database migrations
├── pkg                  public library code
│   ├── accesslog        access log middleware
│   ├── graceful         graceful shutdown of HTTP server
│   ├── log              structured and context-aware logger
│   └── pagination       paginated list
└── testdata             test data scripts
```

The top level directories `cmd`, `internal`, `pkg` are commonly found in other popular Go projects, as explained in
[Standard Go Project Layout](https://github.com/golang-standards/project-layout).

Within `internal` and `pkg`, packages are structured by features in order to achieve the so-called
[screaming architecture](https://blog.cleancoder.com/uncle-bob/2011/09/30/Screaming-Architecture.html). For example, 
the `connection` directory contains the application logic related with the connection feature. 

Within each feature package, code are organized in layers (API, service, repository), following the dependency guidelines
as described in the [clean architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).


## Common Development Tasks

This section describes some common development tasks using this restful client.

### Implementing a New Feature

Implementing a new feature typically involves the following steps:

1. Develop the service that implements the business logic supporting the feature. Please refer to `internal/connection/service.go` as an example.
2. Develop the RESTful API exposing the service about the feature. Please refer to `internal/connection/api.go` as an example.
3. Develop the repository that persists the data entities needed by the service. Please refer to `internal/connection/repository.go` as an example.
4. Wire up the above components together by injecting their dependencies in the main function. Please refer to 
   the `connection.RegisterHandlers()` call in `cmd/server/main.go`.

### Working with DB Transactions

It is the responsibility of the service layer to determine whether DB operations should be enclosed in a transaction.
The DB operations implemented by the repository layer should work both with and without a transaction.

You can use `dbcontext.DB.Transactional()` in a service method to enclose multiple repository method calls in
a transaction. For example,

```go
func serviceMethod(ctx context.Context, repo Repository, transactional dbcontext.TransactionFunc) error {
    return transactional(ctx, func(ctx context.Context) error {
        repo.method1(...)
        repo.method2(...)
        return nil
    })
}
```

If needed, you can also enclose method calls of different repositories in a single transaction. The return value
of the function in `transactional` above determines if the transaction should be committed or rolled back.

You can also use `dbcontext.DB.TransactionHandler()` as a middleware to enclose a whole API handler in a transaction.
This is especially useful if an API handler needs to put method calls of multiple services in a transaction.


### Updating Database Schema

The self restful client uses [database migration](https://en.wikipedia.org/wiki/Schema_migration) to manage the changes of the 
database schema over the whole project development phase. The following commands are commonly used with regard to database
schema changes:

```shell
# Execute new migrations made by you or other team members.
# Usually you should run this command each time after you pull new code from the code repo. 
make migrate

# Create a new database migration.
# In the generated `migrations/*.up.sql` file, write the SQL statements that implement the schema changes.
# In the `*.down.sql` file, write the SQL statements that revert the schema changes.
make migrate-new

# Revert the last database migration.
# This is often used when a migration has some issues and needs to be reverted.
make migrate-down

# Clean up the database and rerun the migrations from the very beginning.
# Note that this command will first erase all data and tables in the database, and then
# run all migrations. 
make migrate-reset
```

### Managing Configurations

The application configuration is represented in `internal/config/config.go`. When the application starts,
it loads the configuration from environment variables.
Configurations
specified in environment variables should be named with the `RESTFUL_CLIENT_` prefix and in upper case. When a configuration. 

### Managed apps

There are 2 ways you can setup your app on this service, based on environment variable, or dynamically setting it up through the api interface.

*Environment variable based configuration*

These are the environment variables that you'll need to provide to setup your app.
```
	- RESTFUL_CLIENT_APP_ID : the default self app identifier.
	- RESTFUL_CLIENT_APP_SECRET : the default self app secret.
	- RESTFUL_CLIENT_APP_ENV : the default self app environment.
	- RESTFUL_CLIENT_APP_MESSAGE_NOTIFICATION_URL : the callback url for any incoming messages on the default app.
  - RESTFUL_CLIENT_APP_DL_CODE : the code (get it from the developer portal) used to build dynamic links.
```

*Dynamically created apps*

You can easily create apps through the rest api, check this example:
```sh
curl -X 'POST' 'http://localhost:8080/v1/login' -H 'accept: application/json' -H 'Content-Type: application/json' -d '{ "username": "<user>", "password": "<password>" }'
curl -X 'POST' 'http://localhost:8080/v1/apps' -H 'accept: application/json' -H 'Authorization: Bearer <BEARER TOKEN>' -H 'Content-Type: application/json' -d '{ "id": "<APP_ID>", "secret": "<DEVICE_APP_SECRET>", "name": "<APP_NAME>", "env":"<APP_ENVIRONMENT>", "callback":"<CALLBACK>", "code":"<CODE>" }'
```

## Client generation

This service comes with openapi implementation, so you can generate a SDK on your preferred language, we're using [openapi generator](https://github.com/OpenAPITools/openapi-generator) with a syntax like:
```shell
openapi-generator generate -i docs/swagger.yaml -g ruby -o ../joinself-restful-ruby
```
Visit [OpenAPI Generator](https://github.com/OpenAPITools/openapi-generator) site to check if your language is supported.


## Support

Looking for help? Reach out to us at [support@joinself.zendesk.com](mailto:support@joinself.zendesk.com)

## Contributing

Check out the [Contributing](CONTRIBUTING.md) guidelines.

## License

See [License](LICENSE).
