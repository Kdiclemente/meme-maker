FROM golang:alpine

MAINTAINER Harrison Shoebridge <harrison@theshoebridges.com>

RUN apk update
RUN apk add git

COPY . /go/src/app

WORKDIR /go/src/app

RUN go get -d -v
RUN go install -v
RUN echo "{}" > config.json

CMD ["app"]
