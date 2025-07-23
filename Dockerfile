FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum /app
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

COPY .env/config.yaml /app/config.yaml

RUN apk --no-cache add postgresql-client

CMD ["/app/app"]
