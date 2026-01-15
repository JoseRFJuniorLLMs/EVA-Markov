# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiar go.mod e go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o markov ./cmd/scheduler

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar binário
COPY --from=builder /app/markov .

# Copiar migrations
COPY --from=builder /app/migrations ./migrations

# Expor porta (se necessário para health check)
EXPOSE 8080

# Comando
CMD ["./markov"]
