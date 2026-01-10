FROM golang:1.25-alpine AS builder

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 ensures a static binary (crucial for Alpine)
# GOOS and GOARCH are automatically handled by buildx --platform
RUN CGO_ENABLED=0 go build -v -o /app/server ./cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/server .
COPY --from=builder /usr/src/app/internal/migrations ./internal/migrations

EXPOSE 50051

CMD ["./server"]