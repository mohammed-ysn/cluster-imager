FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /cluster-imager

FROM gcr.io/distroless/static-debian12

COPY --from=builder /cluster-imager /cluster-imager

EXPOSE 8080

CMD ["/cluster-imager"]
