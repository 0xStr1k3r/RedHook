FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY backend/src ./src
RUN go build -o /phishguard ./src

FROM alpine:latest
WORKDIR /app
COPY --from=builder /phishguard .
COPY backend/.env.example .env
EXPOSE 8080
CMD ["./phishguard"]