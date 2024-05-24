FROM golang:1.22.1 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/myapp

# Etapa final
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/myapp /app/myapp
COPY .env /app/.env
ENTRYPOINT ["/app/myapp"]