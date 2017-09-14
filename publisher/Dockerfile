FROM golang:1.9-alpine

RUN apk add --no-cache git

WORKDIR /go/src/rabbitrunr/publisher

COPY . .

RUN go build

CMD ["./publisher"]

