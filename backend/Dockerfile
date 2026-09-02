FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum* ./
COPY . .
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /contact-service ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=builder /contact-service ./contact-service
COPY --from=builder /app/migrations ./migrations

USER appuser
EXPOSE 8080
ENTRYPOINT ["./contact-service"]