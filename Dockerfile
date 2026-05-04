FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY backend/src ./src
RUN go build -o /redhook ./src

FROM alpine:latest
WORKDIR /app
COPY --from=builder /redhook .
COPY backend/.env.example .env
EXPOSE 8080
CMD ["./redhook"]