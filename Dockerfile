FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app

COPY go.mod ./
COPY go.sum* ./

RUN go mod tidy && go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o myapp ./cmd/api/main.go



FROM alpine:latest
RUN apk add --no-cache tzdata ca-certificates
WORKDIR /root/

COPY --from=builder /app/myapp .
COPY --from=builder /app/database/migrations ./database/migrations

EXPOSE 8080

CMD ["./myapp"]