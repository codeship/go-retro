FROM golang

COPY ./ /go/src/github.com/codeship/retro/
WORKDIR /go/src/github.com/codeship/retro

RUN go get ./...
