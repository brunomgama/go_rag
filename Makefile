.PHONY: ingest up down fmt run-api ingest

ingest:
	go run ./cmd/ingest

up:
	docker compose -f deployments/docker/docker-compose.yml up -d

down:
	docker compose -f deployments/docker/docker-compose.yml down

fmt:
	gofmt -w .
	
run-api: 
	go run ./cmd/api

ingest: 
	go run ./cmd/ingest
