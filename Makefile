run-db: stop-db
	docker run --rm -p 54329:5432 -e POSTGRES_PASSWORD=password -d --name ghloc-db postgres:14-alpine

stop-db:
	-docker stop ghloc-db

# DB is optional, if not provided, the service will be run without cache
run:
	JSON_LOGS=false \
	DB_CONN="postgres://postgres:password@localhost:54329/?sslmode=disable" \
	DEBUG_TOKEN="dt" \
	MAX_REPO_SIZE_MB=100 \
	MAX_CONCURRENT_WORK=2 \
		go run ./cmd/server

run-in-docker:
	docker build -t ghloc .
	docker run --rm -p 8080:8080 \
		-e JSON_LOGS=false \
		-e DEBUG_TOKEN="dt" \
		-e MAX_REPO_SIZE_MB=100 \
		-e MAX_CONCURRENT_WORK=2 \
		ghloc

test:
	go build -v ./...
	go test -cover -v -race ./...