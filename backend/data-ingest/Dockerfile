# Dockerfile for data-ingest microservice
FROM golang:1.22.1-alpine

# Set the working directory
WORKDIR /app

# Copy the data-ingest folder's contents to the container's /app/data-ingest directory
COPY . ./

# Copy the pb and proto folders into the container at /app/pb and /app/proto
COPY ../pb ./pb
COPY ../proto ./proto

# Build the data-ingest microservice
WORKDIR /app/data-ingest

# Install dependencies
RUN go mod tidy

# Build the microservice
RUN go build -o main .

# Expose the port that the microservice runs on
EXPOSE 8080

# Expose gRPC port
EXPOSE 50051

# Start the microservice
CMD ["./main"]
