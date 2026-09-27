FROM golang:1.25.5-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git ca-certificates tzdata

COPY event-service/go.mod event-service/go.sum ./
RUN go mod download

COPY event-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /build/app \
    ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /build/app /app

ENTRYPOINT ["/app"]
