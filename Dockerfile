FROM golang:1.16-buster as builder

# Create and change to the app directory.
WORKDIR /app

# Copy local code to the container image.
COPY . ./


# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:buster-slim
# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/main /app/main

# Run the web service on container startup.
CMD ["/app/main"]


#FROM golang:1.14.9-alpine AS builder
#RUN mkdir /build
#ADD go.mod go.sum main.go /build/
#WORKDIR /build
#RUN go mod tidy
#RUN CGO_ENABLED=0 GOOS=linux GARCH=amd64 go build -o main .
#CMD ["/build/main"]