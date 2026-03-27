FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/api /app/api

ENV PORT=3000
ENV DB_PATH=/data/leads.db

EXPOSE 3000

CMD ["/app/api"]