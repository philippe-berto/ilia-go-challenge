FROM golang:1.25.3 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ms-users ./cmd

FROM scratch
WORKDIR /app
COPY --from=build /app/ms-users .
COPY --from=build /app/migrations ./migrations
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./ms-users"]