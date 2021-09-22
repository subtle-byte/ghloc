FROM golang:1.17

WORKDIR /src
ADD . .
RUN CGO_ENABLED=0 go build -o /bin/app cmd/main.go

# FROM alpine
FROM busybox
# FROM scratch

WORKDIR /bin
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /bin/app .
EXPOSE 8080

ENTRYPOINT ["app"]
