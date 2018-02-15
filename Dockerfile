FROM alpine:latest

RUN apk update && apk add git go libc-dev

RUN mkdir -p /gopath/src/karma.run

WORKDIR /gopath/src/karma.run
RUN git clone --branch 0.5-beta.deterministic https://github.com/karmarun/karma.run.git .

WORKDIR /gopath
RUN GOPATH=$(pwd) go get karma.run
RUN mv bin/karma.run /karma.run
RUN rm -rf gopath

WORKDIR /
CMD "/karma.run"
