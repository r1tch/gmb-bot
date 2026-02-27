.PHONY: test build run dev

test:
	go test ./...

build:
	go build -o bin/gmb-bot ./cmd/gmb-bot

run:
	docker compose up --build -d gmb-bot

dev:
	ONE_SHOT=true LOG_LEVEL=debug docker compose up --abort-on-container-exit --build gmb-bot
