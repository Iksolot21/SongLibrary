FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o songlibrary ./cmd/songlibrary/main.go

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/songlibrary ./songlibrary

COPY --from=builder /app/internal/migrations ./internal/migrations

COPY .env ./.env 

RUN chmod +x ./songlibrary

EXPOSE 8080

CMD ["./songlibrary"]