FROM golang:1.19.1-bullseye

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl libsodium-dev && \
    curl https://download.joinself.com/olm/libself-olm_0.1.17_amd64.deb -o /tmp/libself-olm_0.1.17_amd64.deb && \
    curl https://download.joinself.com/omemo/libself-omemo_0.1.3_amd64.deb -o /tmp/libself-omemo_0.1.3_amd64.deb && \
    apt-get install -y --no-install-recommends /tmp/libself-olm_0.1.17_amd64.deb /tmp/libself-omemo_0.1.3_amd64.deb && \
    curl -Lo /tmp/migrate.tar.gz https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz && \
    tar zxf /tmp/migrate.tar.gz -C /tmp && \
    cp /tmp/migrate /usr/local/bin/ && \
    rm -rf /var/lib/apt/lists/* /tmp/*

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN make build

RUN cp /app/cmd/server/entrypoint.sh .

ENTRYPOINT ["./entrypoint.sh"]
