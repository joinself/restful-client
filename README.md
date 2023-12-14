# Go RESTful Joinself Client

This joinself client is designed to act as a proxy between your app and The Self Network. It exposes a RESTful API to interact with the basic features Self has to offer.

> :warning: **Please note this software is currently under active development and is subject to backwards incompatible changes at any time.**

## Getting Started

### Docker compose

This project is provided with an easy setup based on [docker-compose](https://docs.docker.com/compose/). This will allow you to run the api without dealing with environment configurations.

To run the restful client based on docker compose you just need to configure `config/local.yml` file with your data and run:
```
$ docker-compose up
```

### Running it locally

If this is your first time encountering Go, please follow [the instructions](https://golang.org/doc/install) to
install Go on your computer. The kit requires **Go 1.17 or above**.

[Docker](https://www.docker.com/get-started) is also needed if you want to try the kit without setting up your
own database server. The kit requires **Docker 17.05 or higher** for the multi-stage build support.

After installing Go and Docker, run the following commands to start experiencing this restflu client:

```shell
# download the restful self client
git clone https://github.com/joinself/restful-client.git

cd restful-client

# seed the database with some test data
make testdata

# run the RESTful API server
make run

# or run the API server with live reloading, which is useful during development
# requires fswatch (https://github.com/emcrisostomo/fswatch)
make run-live
```

At this point you should be able to access the RESTful client server accessible through at `http://127.0.0.1:8080`. 

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
├── config               configuration files for different environments
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
it loads the configuration from a configuration file as well as environment variables. The path to the configuration 
file is specified via the `-config` command line argument which defaults to `./config/local.yml`. Configurations
specified in environment variables should be named with the `APP_` prefix and in upper case. When a configuration
is specified in both a configuration file and an environment variable, the latter takes precedence. 

The `config` directory contains the configuration files named after different environments. For example,
`config/local.yml` corresponds to the local development environment and is used when running the application 
via `make run`.

Do not keep secrets in the configuration files. Provide them via environment variables instead. For example,
you should provide `Config.DSN` using the `APP_DSN` environment variable. Secrets can be populated from a secret
storage (e.g. HashiCorp Vault) into environment variables in a bootstrap script (e.g. `cmd/server/entryscript.sh`). 

## Deployment

The application can be run as a docker container. You can use `make build-docker` to build the application 
into a docker image. The docker container starts with the `cmd/server/entryscript.sh` script which reads 
the `APP_ENV` environment variable to determine which configuration file to use. For example,
if `APP_ENV` is `qa`, the application will be started with the `config/qa.yml` configuration file.

You can also run `make build` to build an executable binary named `server`. Then start the API server using the following
command,

```shell
./server -config=./config/prod.yml
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