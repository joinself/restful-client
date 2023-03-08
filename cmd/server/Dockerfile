FROM golang:1.19.1-bullseye AS build

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl libsodium-dev && \
    curl https://download.joinself.com/olm/libself-olm_0.1.17_amd64.deb -o /tmp/libself-olm_0.1.17_amd64.deb && \
    curl https://download.joinself.com/omemo/libself-omemo_0.1.3_amd64.deb -o /tmp/libself-omemo_0.1.3_amd64.deb && \
    apt-get install -y --no-install-recommends /tmp/libself-olm_0.1.17_amd64.deb /tmp/libself-omemo_0.1.3_amd64.deb && \
    rm -rf /var/lib/apt/lists/* /tmp/*
    
RUN apt-get install -y curl \
            git \
            bash \
            make \
            ca-certificates && \

# install migrate which will be used by entrypoint.sh to perform DB migration
ADD https://github.com/golang-migrate/migrate/releases/download/v4.7.1/migrate.linux-amd64.tar.gz /tmp
RUN tar -xzf /tmp/migrate.linux-amd64.tar.gz -C /usr/local/bin && mv /usr/local/bin/migrate.linux-amd64 /usr/local/bin/migrate

WORKDIR /app

# copy module files first so that they don't need to be downloaded again if no change
COPY go.* ./
RUN go mod download
RUN go mod verify

# copy source files and build the binary
COPY . .
RUN make build


FROM alpine:latest
RUN apk --no-cache add ca-certificates bash
RUN mkdir -p /var/log/app
WORKDIR /app/
COPY --from=build /usr/local/bin/migrate /usr/local/bin
COPY --from=build /app/migrations ./migrations/
COPY --from=build /app/server .
COPY --from=build /app/cmd/server/entrypoint.sh .
COPY --from=build /app/config/*.yml ./config/
RUN ls -la
ENTRYPOINT ["./entrypoint.sh"]
