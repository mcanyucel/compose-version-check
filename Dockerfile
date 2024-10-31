FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o compose-checker

FROM alpine:3.19

WORKDIR /app
# Copy the binary
COPY --from=builder /build/compose-checker /app/
# Copy example config
COPY config.yaml.example /app/config.yaml.example

# Create watch directory - this is where docker-compose files will be mounted
RUN mkdir /watch

# The application will automatically detect it's running in Docker
# and prepend /watch to the paths in the config file
ENTRYPOINT ["/app/compose-checker"]
CMD ["-config", "/app/config.yaml"]