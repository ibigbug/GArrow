FROM golang:1.8

MAINTAINER Yuwei Ba <akabyw@gmail.com>

RUN go get github.com/ibigbug/GArrow/cmd/garrow

CMD ["garrow"]
