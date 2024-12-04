FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o compose-checker compose-version-check.go

FROM mcuadros/ofelia:latest

WORKDIR /app
# Copy the binary
COPY --from=builder /build/compose-checker /app/

# Create required directories
RUN mkdir /watch && \
    mkdir -p /etc/ofelia

# Create startup script that will create config at runtime
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'echo "Setting up scheduler to run every ${CHECK_INTERVAL:-6h}"' >> /app/start.sh && \
    echo 'echo "[job-local \"compose-checker\"]" > /etc/ofelia/config.ini' >> /app/start.sh && \
    echo 'echo "schedule = @every ${CHECK_INTERVAL:-6h}" >> /etc/ofelia/config.ini' >> /app/start.sh && \
    echo 'echo "command = /app/compose-checker -config /app/config.yaml" >> /etc/ofelia/config.ini' >> /app/start.sh && \
    echo 'echo "Running initial check..."' >> /app/start.sh && \
    echo '/app/compose-checker -config /app/config.yaml' >> /app/start.sh && \
    echo 'echo "Starting scheduler..."' >> /app/start.sh && \
    echo 'exec ofelia daemon --config=/etc/ofelia/config.ini' >> /app/start.sh && \
    chmod +x /app/start.sh

ENTRYPOINT ["/app/start.sh"]