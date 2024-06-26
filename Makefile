MODULE = $(shell go list -m)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo "1.0.0")
PACKAGES := $(shell go list ./... | grep -v /vendor/)
LDFLAGS := -ldflags "-X main.Version=${VERSION}"

APP_DSN = "sqlite3://${RESTFUL_CLIENT_STORAGE_DIR}/client.db"
APP_DSN_TEST = "sqlite3://${RESTFUL_CLIENT_STORAGE_DIR}/client-test.db"
MIGRATE := migrate -path=./migrations/ -database "$(APP_DSN)"
MIGRATE_TEST := migrate -path=./migrations/ -database "$(APP_DSN_TEST)"

PID_FILE := './.pid'
FSWATCH_FILE := './fswatch.cfg'

.PHONY: default
default: help

# generate help info from comments: thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## help information about make commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## run unit tests
	go test -p=1 -cover -covermode=count -coverprofile=coverage.out ./...

.PHONY: unit-test-cover
test-cover: ## run unit tests and show test coverage information
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES), \
		go test -p=1 -cover -covermode=count -coverprofile=coverage.out ${pkg}; \
		tail -n +2 coverage.out >> coverage-all.out;)

	go tool cover -html=coverage-all.out

.PHONY: run
run: ## run the API server
	go run ${LDFLAGS} cmd/server/main.go

.PHONY: run-restart
run-restart: ## restart the API server
	@pkill -P `cat $(PID_FILE)` || true
	@printf '%*s\n' "80" '' | tr ' ' -
	@echo "Source file changed. Restarting server..."
	@go run ${LDFLAGS} cmd/server/main.go & echo $$! > $(PID_FILE)
	@printf '%*s\n' "80" '' | tr ' ' -

run-live: ## run the API server with live reload support (requires fswatch)
	@go run ${LDFLAGS} cmd/server/main.go & echo $$! > $(PID_FILE)
	@fswatch -x -o --event Created --event Updated --event Renamed -r internal pkg cmd config | xargs -n1 -I {} make run-restart

.PHONY: build
build:  ## build the API server binary
	go build ${LDFLAGS} -a -o server $(MODULE)/cmd/server

.PHONY: build-docker
build-docker: ## build the API server as a docker image
	docker build -f=cmd/server/Dockerfile -t server .

.PHONY: clean
clean: ## remove temporary files
	rm -rf server coverage.out coverage-all.out

.PHONY: version
version: ## display the version of the API server
	@echo $(VERSION)

.PHONY: testdata
testdata: ## populate the database with test data
	make migrate-reset
	@echo "Populating test data..."
	@sqlite3 APP_DSN_TEST < /testdata/testdata.sql

.PHONY: lint
lint: ## run golint on all Go package
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		wget -O - -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(go env GOPATH)/bin v1.21.0; \
	fi
	golangci-lint run -v ./...

.PHONY: fmt
fmt: ## run "go fmt" on all Go packages
	@go fmt $(PACKAGES)

.PHONY: migrate
migrate: ## run all new database migrations
	@echo "Running all new database migrations..."
	@echo " -> $(MIGRATE) up"
	@$(MIGRATE) up

.PHONY: migrate-down
migrate-down: ## revert database to the last migration step
	@echo "Reverting database to the last migration step..."
	@$(MIGRATE) down 1

.PHONY: migrate-new
migrate-new: ## create a new database migration
	@read -p "Enter the name of the new migration: " name; \
	$(MIGRATE) create -ext sql -dir ./migrations/ $${name// /_}

.PHONY: migrate-reset
migrate-reset: ## reset database and re-run all migrations
	@echo "Resetting database..."
	@$(MIGRATE) drop
	@echo "Running all database migrations..."
	@$(MIGRATE) up

.PHONY: migrate-test
migrate-test: ## run all new database migrations
	@echo "Running all new database migrations..."
	@echo " -> $(MIGRATE_TEST) up"
	@$(MIGRATE_TEST) up

.PHONY: migrate-test-reset
migrate-test-reset: ## reset database and re-run all migrations
	@echo "Resetting test database..."
	@$(MIGRATE_TEST) drop
	@echo "Running all test database migrations..."
	@$(MIGRATE_TEST) up

.PHONY: gen-openapi
gen-openapi: ## generates openapi documentation under /docs folder
	swag init -d cmd/server/,internal/healthcheck/,internal/auth,internal/connection,internal/fact,internal/message,internal/entity,internal/request,internal/app,internal/notification,internal/account,pkg/response,internal/apikey
