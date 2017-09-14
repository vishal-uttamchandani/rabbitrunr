FROM golang:1.9-alpine

RUN apk add --no-cache git

WORKDIR /go/src/rabbitrunr/consumer

COPY . .

RUN go build

CMD ["./consumer"]

