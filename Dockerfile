FROM golang:latest as builder

WORKDIR /go/activemq-archiver
COPY . /go/activemq-archiver

RUN go version

RUN make archiver

FROM scratch

COPY --from=builder /go/activemq-archiver/build/activemq-archiver /
CMD ["/activemq-archiver"]

