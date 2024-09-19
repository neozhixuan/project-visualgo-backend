# Dockerfile for trading-algo microservice
FROM golang:1.22.1-alpine

# Set the working directory
WORKDIR /app

# Copy the microservice code into the container
COPY ./trading-algo ./trading-algo

# Copy the /pb and /proto folders into the working directory
COPY ../pb ./pb
COPY ../proto ./proto

# Build the trading-algo microservice
WORKDIR /app/trading-algo

# Install dependencies
RUN go mod tidy

# Build the microservice
RUN go build -o main .

# Set environment variables
ENV STAGE=production

# Expose the port that the microservice runs on
EXPOSE 8090

# Start the microservice
CMD ["./main"]
