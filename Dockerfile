# syntax=docker/dockerfile:1

FROM golang:1.27-bookworm AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/server ./cmd/server

FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        libreoffice-writer \
        unoconv \
        fonts-dejavu-core \
        fonts-liberation2 \
        supervisor \
        curl \
        ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/server ./server
COPY templates ./templates
COPY docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
RUN mkdir -p output/Docx output/PDF

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/usr/bin/supervisord", "-n", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
