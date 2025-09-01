FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -o finance-bot main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/finance-bot .
COPY --from=builder /app/finance.db ./finance.db
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/internal/i18n/locales ./internal/i18n/locales
COPY --from=builder /app/internal/i18n/locales ./locales

EXPOSE 8080

CMD ["./finance-bot"]