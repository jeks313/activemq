FROM golang:latest as builder

WORKDIR /go
COPY .. /go
WORKDIR /go/archiver

RUN make build

FROM scratch

COPY --from=builder /go/build/activemq-archive /
CMD ["/activemq-archive"]

