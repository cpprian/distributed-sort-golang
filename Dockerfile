FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o edcs-go

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/edcs .

CMD ["./edcs-go"]
