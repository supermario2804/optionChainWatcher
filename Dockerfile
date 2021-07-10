FROM golang:1.14.9-alpine AS builder
RUN mkdir /build
ADD go.mod go.sum main.go /build/
WORKDIR /build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GARCH=amd64 go build -o main .
CMD ["./main"]