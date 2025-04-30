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

# Create a directory for persistent data (like SSH keys)
RUN mkdir /data

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# ─── install terminfo & tput ───────────────────────────────────────────────────
# ncurses provides the `tput` utility; ncurses-terminfo brings in /usr/share/terminfo
RUN apk add --no-cache ncurses ncurses-terminfo
# ────────────────────────────────────────────────────────────────────────────────

# Ensure that the color is set to 256
ENV TERM=xterm-256color

# Declare /data as a volume for persistent key storage
VOLUME /data

# Expose any necessary ports (if your app listens on a port)
# EXPOSE 8080
# Default SSH port for this application is 22 (can be overridden to 23234 with DEV_MODE=true)
EXPOSE 22

# Command to run the executable
CMD ["./main"]
