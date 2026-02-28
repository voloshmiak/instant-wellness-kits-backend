FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy && \
    go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./src/cmd/main/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/src/migrations ./src/migrations
COPY --from=builder /app/tax_rates.csv ./tax_rates.csv

CMD ["./main"]