FROM golang:1.21

# Working directory
WORKDIR /app
COPY . /app
RUN go build -o /app/main cmd/main.go
CMD ["/app/main"]
