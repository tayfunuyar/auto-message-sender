
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod ./

RUN go mod download

COPY . .

RUN swag init -g cmd/api/main.go -o ./docs

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

FROM alpine:latest

WORKDIR /app

RUN mkdir -p /app/config

COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs

ENV APP_ENV=prod

EXPOSE 8080

CMD ["./main"] 