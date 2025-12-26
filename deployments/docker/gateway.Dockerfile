FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git protobuf protobuf-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /gateway ./cmd/gateway

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=builder /gateway /gateway

EXPOSE 8080

ENTRYPOINT ["/gateway"]
