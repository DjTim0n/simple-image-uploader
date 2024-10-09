FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o app

RUN mkdir -p ./images

RUN mkdir -p ./db

EXPOSE 4000

CMD ["./app"]
