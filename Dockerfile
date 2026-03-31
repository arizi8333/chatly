# Build Stage
FROM golang:1.25.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Final Stage
FROM alpine:latest
WORKDIR /app/
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/static ./static
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/.env.example ./.env

EXPOSE 8080
CMD ["./main"]
