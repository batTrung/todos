FROM golang:1.20

WORKDIR /app

COPY . .
RUN go mod download

COPY *.go ./

RUN go build -o /connection-pool

CMD ["/connection-pool"]
