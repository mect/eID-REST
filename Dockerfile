FROM go:1.18 as build

COPY ./ /go/src/github.com/mect/eID-REST
WORKDIR /go/src/github.com/mect/eID-REST

RUN go build ./cmd/eid-rest

FROM ubuntu:20.04

RUN apt-gett update && \
    apt-get install -y opensc

COPY --from=build /go/src/github.com/mect/eID-REST/eid-rest /usr/bin/eid-rest

ENTRYPOINT [ "/usr/bin/eid-rest" ]
CMD [ "serve" ]