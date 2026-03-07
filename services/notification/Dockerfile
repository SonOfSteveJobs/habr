FROM golang:1.26.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.24.3

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./services/notification/cmd

FROM alpine:3.23.3

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/server .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/migrations/notification ./migrations

RUN chown -R appuser:appgroup /app

USER appuser

CMD ["./server"]
