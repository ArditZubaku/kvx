# Stage 0: Install dependencies
FROM golang:1.26.4-alpine3.23 AS deps

WORKDIR /app

COPY go.mod ./

RUN go mod download

# Stage 1: Test stage
FROM deps AS test

COPY . .

# Run tests with no cache
RUN go test -v -count=1 ./...

# Stage 2: Build the application
FROM deps AS builder

WORKDIR /app

COPY . .

# Enable them if you need them
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build -ldflags="-w -s" -o main .

# Final stage: Run the application
FROM debian:bookworm-slim

WORKDIR /app

# Create a non-root user and group
RUN groupadd -r appuser && useradd -r -g appuser appuser

# Copy the built application
COPY --from=builder /app/main .

# Change ownership of the /app directory itself, not just the binary
# We need it in order to be able to write the AOF file
RUN chown -R appuser:appuser /app

# Switch to the non-root user
USER appuser

CMD ["./main"]
