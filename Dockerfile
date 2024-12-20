# Используем минимальный базовый образ для Go
FROM golang:1.20-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git

# Задаём рабочую директорию
WORKDIR /app

# Копируем файлы приложения
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Запуск тестов (если нужно)
RUN go test -v ./... -coverprofile=coverage.out

# Собираем бинарный файл
RUN go build -o main .

# Создаём финальный минимальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY .env .

ENV TELEGRAM_BOT_TOKEN=""

CMD ["./main"]
