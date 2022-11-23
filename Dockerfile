FROM golang:1.19 as builder

ADD . github.com/gideonw/peltr
WORKDIR /go/github.com/gideonw/peltr

RUN go build -o /go/bin/peltr main.go 

FROM ubuntu:latest
RUN apt-get update && apt-get install curl -y
COPY --from=builder /go/bin/peltr /bin/peltr

ENTRYPOINT ["/bin/peltr"]