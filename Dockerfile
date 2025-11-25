FROM golang:1.25.4-alpine3.21 AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o rezaserver

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/rezaserver .
CMD ["./rezaserver"]

