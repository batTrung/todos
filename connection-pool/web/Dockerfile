FROM golang:1.20

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build

CMD ["go", "run", "main.go"]
