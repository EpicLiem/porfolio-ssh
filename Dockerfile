# ---- Builder Stage ----
# Use the official Go image corresponding to the version in go.mod
FROM golang:1.24.2-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
# CGO_ENABLED=0 ensures a static binary (useful for minimal images)
# -ldflags "-s -w" strips debugging information to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

# ---- Final Stage ----
# Use a minimal base image
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose any necessary ports (if your app listens on a port)
# EXPOSE 8080

# Command to run the executable
CMD ["./main"] 