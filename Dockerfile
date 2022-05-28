FROM golang:1.17

WORKDIR /src
ADD . .
RUN CGO_ENABLED=0 go build -v -ldflags "-s -X main.buildTime=`date --iso-8601=seconds -u`" -o /bin/app ./cmd/server

# FROM alpine
FROM busybox
# FROM scratch

WORKDIR /app
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /bin/app .
COPY --from=0 /src/migrations ./migrations
EXPOSE 8080

ENTRYPOINT ["/app/app"]
