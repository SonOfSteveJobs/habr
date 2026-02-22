FROM golang:1.26.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./services/auth/cmd

FROM alpine:3.23.3

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/server .
COPY --from=builder /app/migrations/auth ./migrations

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 50051

CMD ["./server"]
