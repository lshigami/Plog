# ----Build FrontEnd----
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend/ ./
RUN npm run build


# ----- Build Stage -----
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/server ./cmd/server/main.go

# ----- Runtime Stage -----
FROM alpine:3.19
WORKDIR /app

RUN apk add --no-cache postgresql-client

RUN wget https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz && \
    tar -xvzf migrate.linux-amd64.tar.gz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate && \
    rm migrate.linux-amd64.tar.gz

COPY --from=builder /app/server /app/server

COPY --from=frontend-builder /app/frontend/build /app/static

COPY internal/db/migrations /app/migrations

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]