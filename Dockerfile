
FROM golang:latest as builder
ENV CGO_ENABLED=0
RUN go get -v "github.com/tsaikd/gobuilder"
ADD . /go/src/github.com/tsaikd/golasticdump
WORKDIR /go/src/github.com/tsaikd/golasticdump
RUN gobuilder --check --test --all

FROM alpine:3.7
MAINTAINER tsaikd <tsaikd@gmail.com>
COPY --from=builder /go/src/github.com/tsaikd/golasticdump/golasticdump /usr/local/bin/golasticdump
ENTRYPOINT ["golasticdump"]
