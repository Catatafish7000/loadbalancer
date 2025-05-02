
FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/loadbalancer .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/loadbalancer .
COPY config.yaml .


#RUN chmod +x /app/loadbalancer

EXPOSE 8080
CMD ["./loadbalancer"]