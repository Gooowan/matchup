.PHONY: gen up down deps

SHELL := /bin/sh

gen:
	go run cmd/schema-gen/main.go

up: gen
	docker compose up -d --build

down:
	docker compose down

deps:
	go mod download
	go mod tidy
