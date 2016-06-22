FROM golang

#RUN mkdir -p /opt/go/src/github.com/codeship/retro
COPY ./ /go/src/github.com/codeship/retro/
WORKDIR /go/src/github.com/codeship/retro


