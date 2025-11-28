FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o queue-core ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/queue-core .
CMD ["./queue-core"]