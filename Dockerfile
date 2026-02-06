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

# Build arguments for OSS configuration
ARG OSS_ENDPOINT
ARG OSS_ACCESS_KEY_ID
ARG OSS_ACCESS_KEY_SECRET
ARG OSS_BUCKET

COPY --from=builder /out/sparkx-service /app/sparkx-service
COPY --from=builder /app/etc /app/etc

# Replace environment variables in config file
RUN sed -i "s|\\${OSS_ENDPOINT}|${OSS_ENDPOINT}|g; s|\\${OSS_ACCESS_KEY_ID}|${OSS_ACCESS_KEY_ID}|g; s|\\${OSS_ACCESS_KEY_SECRET}|${OSS_ACCESS_KEY_SECRET}|g; s|\\${OSS_BUCKET}|${OSS_BUCKET}|g" /app/etc/sparkx-api.yaml

USER app

EXPOSE 8890

ENTRYPOINT ["/app/sparkx-service"]
CMD ["-f", "/app/etc/sparkx-api.yaml"]
