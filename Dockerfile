FROM alpine:latest

RUN apk update && apk add git go libc-dev

RUN mkdir gopath
WORKDIR gopath
RUN GOPATH=$(pwd) go get github.com/karmarun/karma.run
RUN mv bin/karma.run /karma.run
RUN rm -rf gopath

WORKDIR /
CMD "/karma.run"
