FROM golang:1.20-alpine

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum .
RUN go mod download

COPY . .

# Build
RUN go build -o cluster-imager main.go

EXPOSE 8080

CMD ["./cluster-imager"]
