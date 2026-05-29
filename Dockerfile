FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o gocars-api ./cmd/api/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gocars-api .

EXPOSE 9000

CMD ["./gocars-api"]
