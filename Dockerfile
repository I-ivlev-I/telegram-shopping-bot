# syntax=docker/dockerfile:1

FROM golang:1.20-alpine AS base
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base AS test
RUN go test -v ./... -coverprofile=coverage.out

FROM base AS builder
RUN CGO_ENABLED=0 go build -ldflags='-s -w' -o main .

FROM alpine:latest AS runtime
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
