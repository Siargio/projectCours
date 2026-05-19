FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o shortener ./cmd/shortener

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/shortener .
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./shortener"]