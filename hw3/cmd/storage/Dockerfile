FROM golang:1.26-alpine
WORKDIR /app

COPY . .
RUN go build -o /app/exe ./cmd/storage

CMD ["/app/exe"]
