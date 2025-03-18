FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Изменяем путь сборки, чтобы файл оказался в корне
RUN CGO_ENABLED=0 go build -o songlibrary ./cmd/songlibrary/main.go

FROM alpine:3.18

WORKDIR /app

# Добавляем необходимые зависимости для работы Go-приложений
RUN apk add --no-cache ca-certificates tzdata

# Копируем бинарный файл из корня /app
COPY --from=builder /app/songlibrary ./songlibrary

# Копируем файлы миграций
COPY --from=builder /app/internal/migrations ./internal/migrations

COPY .env ./.env 

# Убедимся, что файл имеет права на выполнение
RUN chmod +x ./songlibrary

EXPOSE 8080

CMD ["./songlibrary"]