FROM golang:1.19.1-bullseye

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    curl -Lo /tmp/self-omemo.deb https://github.com/joinself/self-omemo/releases/download/0.5.0/self-omemo_0.5.0_amd64.deb && \
    apt-get install -y --no-install-recommends /tmp/self-omemo.deb && \
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
