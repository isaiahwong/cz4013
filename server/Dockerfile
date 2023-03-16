# Build stage
FROM golang:1.18-alpine AS build

WORKDIR /app

# Copy the go.mod and go.sum files into the container
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the binary with Go
RUN GOOS=linux GOARCH=amd64 go build -o flightsys ./cmd/main.go

# Runtime stage
FROM alpine:3.14

WORKDIR /app

# Copy the binary from the build stage into the runtime stage
COPY --from=build /app/flightsys /app


# Expose the port that the binary listens on
EXPOSE 8080

COPY ./flights.csv .

RUN chmod +x ./flightsys

# Start the binary when the container starts
CMD ["./flightsys", "-deadline", "3", "-semantic", "1", "-loss", "50", "-port", "8080"]

