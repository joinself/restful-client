FROM golang:1.19.1-bullseye AS builder

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    curl -Lo /tmp/self-omemo.deb https://github.com/joinself/self-omemo/releases/download/0.5.0/self-omemo_0.5.0_amd64.deb && \
    apt-get install -y --no-install-recommends /tmp/self-omemo.deb && \
    curl -Lo /tmp/migrate.tar.gz https://download.joinself.com/golang-migrate/migrate-sqlite3-4.16.2.tar.gz && \
    tar -zxf /tmp/migrate.tar.gz -C /tmp

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o restful-client /build/cmd/server


FROM debian:bullseye-20231120-slim

COPY --from=builder /build/restful-client /srv
COPY --from=builder /build/docker/entrypoint.sh /srv
COPY --from=builder /build/migrations /srv/migrations
COPY --from=builder /tmp/migrate /usr/local/bin

ENTRYPOINT ["/srv/entrypoint.sh"]
