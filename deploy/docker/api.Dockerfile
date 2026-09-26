FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/fastimg-api .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S fastimg \
    && adduser -S -G fastimg -h /app fastimg \
    && mkdir -p /app/storage/fastimg /app/storage/logs \
    && chown -R fastimg:fastimg /app

WORKDIR /app
COPY --from=builder /out/fastimg-api /app/fastimg-api
COPY backend/public/ /app/public/
COPY backend/resources/ /app/resources/

RUN chown -R fastimg:fastimg /app

USER fastimg
ENV APP_HOST=0.0.0.0
ENV APP_PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --start-period=20s --retries=5 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/api/v1/discovery/status || exit 1

ENTRYPOINT ["/app/fastimg-api"]
