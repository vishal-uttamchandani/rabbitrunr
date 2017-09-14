FROM golang:1.9-alpine

#RUN apk add --no-cache curl
RUN apk add --no-cache git

WORKDIR /go/src/rabbitrunr/runr

COPY . .

RUN go build

EXPOSE 5000

CMD ["./runr"]




