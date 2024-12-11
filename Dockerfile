FROM golang:1.23-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum .

RUN go mod download && go mod verify

COPY . .

RUN go build -o main ./cmd/app/main.go

FROM alpine:latest

COPY . .

COPY --from=builder /app/main .

EXPOSE 3000

CMD ["./main"]
