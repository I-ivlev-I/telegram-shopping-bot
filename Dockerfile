# Этап 1: сборка
FROM golang:1.23.4-alpine3.21 AS builder

# Установка необходимых пакетов
RUN apk add --no-cache git

# Установка рабочей директории
WORKDIR /app

# Копируем файлы go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Установка зависимостей
RUN go mod download

# Копируем оставшиеся файлы проекта
COPY . .

# Сборка бинарного файла
RUN go build -o main

# Этап 2: финальное изображение
FROM alpine:3.21

# Установка рабочей директории
WORKDIR /app

# Копируем только бинарный файл из этапа сборки
COPY --from=builder /app/main .

# Указываем команду запуска
CMD ["./main"]