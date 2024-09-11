FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o app

RUN mkdir -p ./images

EXPOSE 4000

CMD ["./app"]