run-db: stop-db
	docker run --rm -p 54329:5432 -e POSTGRES_PASSWORD=password -d --name ghloc-db postgres:14-alpine

stop-db:
	-docker stop ghloc-db

run:
	DB_CONN="postgres://postgres:password@localhost:54329/?sslmode=disable" \
	DEBUG_TOKEN="" \
		go run cmd/server/main.go

test:
	go build -v ./...
	go test -cover -v -race ./...