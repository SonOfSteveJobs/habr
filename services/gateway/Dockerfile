FROM golang:1.26.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/bufbuild/buf/cmd/buf@v1.65.0 && \
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1

COPY . .

RUN cd proto && buf generate
RUN oapi-codegen -generate chi-server,models,spec -package gatewayv1 -o pkg/gen/gateway/v1/gateway.gen.go api/gateway/v1/gateway.yaml
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./services/gateway/cmd

FROM alpine:3.23.3

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/server .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./server"]
