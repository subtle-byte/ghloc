run-db: stop-db
	docker run --rm -p 54329:5432 -e POSTGRES_PASSWORD=password -d --name ghloc-db postgres:14-alpine

stop-db:
	-docker stop ghloc-db

# DB is optional, if not provided, the service will be run without cache
run:
	DB_CONN="postgres://postgres:password@localhost:54329/?sslmode=disable" \
	DEBUG_TOKEN="" \
		go run cmd/server/main.go

run-in-docker:
	docker build -t ghloc .
	docker run --rm -p 8080:8080 -e DEBUG_TOKEN="" ghloc

test:
	go build -v ./...
	go test -cover -v -race ./...