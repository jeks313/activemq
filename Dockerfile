FROM golang:latest as builder

WORKDIR /go/activemq-archiver
COPY . /go/activemq-archiver

RUN go version

RUN make archiver

FROM alpine:latest

RUN mkdir -p /var/lib/activemq-archive && chown nobody:nogroup /var/lib/activemq-archive
USER nobody
COPY --from=builder /go/activemq-archiver/build/activemq-archiver /
CMD ["/activemq-archiver"]
