# Use the official Go 1.20 Alpine-based image for building
FROM golang:1.20-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the project files into the working directory
COPY . .

# Install dependencies (if there's a go.mod file)
RUN go mod tidy

# Build the application with optimization for Linux Alpine
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o postback-catcher

# Use the Alpine base image for the final container
FROM alpine:latest

# Set the working directory
WORKDIR /app

RUN mkdir data && chmod -R 777 data

# Copy the executable from the builder container
COPY --from=builder /app/postback-catcher /app/postback-catcher

# Expose port 8081
EXPOSE 8081

# Run the application
CMD [ "./postback-catcher" ]
