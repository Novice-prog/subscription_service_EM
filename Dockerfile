
FROM golang:1.24-alpine AS builder

# Обновляем индексы и ставим сертификаты (git необязательно, если все зависимости в proxy)
RUN apk update && apk add --no-cache ca-certificates

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник из корня (где лежит main.go)
RUN go build -o app .

# --- Этап минимального runtime ---
FROM alpine:latest

# Устанавливаем сертификаты
RUN apk update && apk add --no-cache ca-certificates

WORKDIR /app

# Копируем бинарник и миграции
COPY --from=builder /app/app .
COPY --from=builder /app/migrations ./migrations

# Пробрасываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./app"]
