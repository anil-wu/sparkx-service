FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/sparkx-service .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && update-ca-certificates

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/sparkx-service /app/sparkx-service
COPY --from=builder /app/etc /app/etc

USER app

EXPOSE 8890

ENTRYPOINT ["/app/sparkx-service"]
CMD ["-f", "/app/etc/sparkx-api.yaml"]
