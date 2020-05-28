FROM golang:1.14.1

WORKDIR /go/src/algorithm2.0/

RUN go get -d -v github.com/fluent/fluent-logger-golang/fluent

COPY constants ./constants
COPY logging ./logging
COPY simulation ./simulation
COPY types ./types
COPY util ./util
COPY vehicle ./vehicle
COPY main.go .

RUN go build -a -o app .



FROM ubuntu:latest

WORKDIR /root/

COPY --from=0 /go/src/algorithm2.0/app .

CMD ["./app"]
