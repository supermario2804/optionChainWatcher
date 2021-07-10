FROM golang:1.14.9-alpine AS builder
ADD go.mod go.sum main.go ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GARCH=amd64 go build -o main .
CMD ["./main"]