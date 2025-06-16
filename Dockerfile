FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/edcs-sort ./bin

FROM debian:bookworm-slim

COPY --from=builder /bin/edcs-sort /usr/local/bin/edcs-sort

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/local/bin/edcs-sort"]

EXPOSE 8080

CMD ["./edcs-sort"]