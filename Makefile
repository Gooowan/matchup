.PHONY: gen up down deps seed-profiles delete-user observability-up benchmark

SHELL := /bin/sh

gen:
	go run cmd/schema-gen/main.go

up: gen
	cd services/frontend && pnpm install && pnpm build
	docker compose up -d --build
	docker compose restart web

down:
	docker compose down

deps:
	go mod download
	go mod tidy

# Seed mock dancer profiles into the database.
# Usage: make seed-profiles              (creates 30 profiles)
#        make seed-profiles COUNT=50     (creates 50 profiles)
#        make seed-profiles COUNT=20 ASSIGN_CLUBS=false
seed-profiles:
	go run ./cmd/seed-profiles \
		--count=$(or $(COUNT),30) \
		--assign-clubs=$(or $(ASSIGN_CLUBS),true)

# Fully delete a user and all their data from the database.
# Usage: make delete-user EMAIL=foo@example.com
#        make delete-user ID=<uuid>
#        make delete-user EMAIL=foo@example.com YES=true   (skip prompt)
delete-user:
	@if [ -z "$(EMAIL)$(ID)" ]; then \
		echo "Usage: make delete-user EMAIL=foo@bar.com   # or ID=<uuid>"; exit 1; \
	fi
	go run ./cmd/delete-user \
		$(if $(EMAIL),--email=$(EMAIL),--id=$(ID)) \
		$(if $(YES),--yes)

observability-up:
	chmod +x scripts/observability-up.sh
	./scripts/observability-up.sh $(if $(PROD),--prod,)

benchmark:
	chmod +x scripts/benchmark.sh
	@if [ -z "$(EMAIL)$(PASS)" ]; then \
		echo "Usage: make benchmark EMAIL=user@example.com PASS=secret [API=https://...] [DURATION=2m]"; exit 1; \
	fi
	./scripts/benchmark.sh $(EMAIL) $(PASS) $(or $(API),http://localhost:8000) $(or $(DURATION),2m)
