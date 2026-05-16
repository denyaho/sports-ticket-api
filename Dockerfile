FROM golang:1.24.5-alpine

# Install air
RUN apk add --update --no-cache ca-certificates git && \
    go install github.com/air-verse/air@v1.62.0

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Run air
CMD ["air"]