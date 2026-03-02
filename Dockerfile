FROM golang:1.25.0-alpine AS builder
ARG SERVICE_NAME=api
ARG ARCHITECTURE=arm64

ENV GOMODCACHE=/go-cache/mod
ENV GOCACHE=/go-cache/build

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go-cache \
    go mod download

COPY ./shared ./shared
COPY ./modules ./modules
COPY ./services/$SERVICE_NAME ./services/$SERVICE_NAME

RUN --mount=type=cache,target=/go-cache \
    CGO_ENABLED=0 GOOS=linux GOARCH=$ARCHITECTURE \
    go build -ldflags="-w -s" -trimpath -o service ./services/$SERVICE_NAME

FROM alpine:3

RUN apk --no-cache add ca-certificates wget
WORKDIR /app

COPY --from=builder /app/service .

EXPOSE 8000

CMD ["./service"] 
