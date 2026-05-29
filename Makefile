.PHONY: run tidy up

run:
	go run ./cmd/api

tidy:
	go mod tidy

up:
	docker compose up -d
