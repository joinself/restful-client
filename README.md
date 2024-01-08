# Self RESTful Client

This client is designed to act as a proxy between your app and the Self network. It exposes a RESTful API to interact with the basic features Self has to offer.

> :warning: **Please note this software is currently under active development and is subject to backwards incompatible changes at any time.**

## Getting Started

There are a few different ways to run this service depending on what it is you want to do. The easiest method is to use Docker Compose with the example configuration from the repository. If you're looking to contribute back to the project then the `Build from source` section is for you.

### Docker Compose

An example `docker-compose.yml` can be found in the `docker` directory at the root of this project.

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

## RESTful API

For a full list of endpoints see [Swagger](https://editor.swagger.io/?url=https://raw.githubusercontent.com/joinself/restful-client/main/docs/swagger.json).

### Examples

API authentication:
```bash
curl -X 'POST' 'http://localhost:8080/v1/login' -H 'accept: application/json' -H 'Content-Type: application/json' -d '{ "username": "<user>", "password": "<password>" }'
```

Create an application:
```bash
curl -X 'POST' 'http://localhost:8080/v1/apps' -H 'accept: application/json' -H 'Authorization: Bearer <BEARER TOKEN>' -H 'Content-Type: application/json' -d '{ "id": "<APP_ID>", "secret": "<DEVICE_APP_SECRET>", "name": "<APP_NAME>", "env":"<APP_ENVIRONMENT>", "callback":"<CALLBACK>", "code":"<CODE>" }'
```

## OpenAPI

This service follows the OpenAPI specification. API client libraries (SDKs), server stubs, documentation and configuration can be automatically built for your preferred language using [openapi-generator](https://github.com/OpenAPITools/openapi-generator).


```bash
openapi-generator generate -i docs/swagger.yaml -g ruby -o self-restful-client-ruby
```

## Support

Looking for help? Reach out to us at [support@joinself.com](mailto:support@joinself.com)

## Contributing

See [Contributing](CONTRIBUTING.md).

## License

See [License](LICENSE).
