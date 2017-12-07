FROM alpine:latest

RUN apk update && apk add git go libc-dev

RUN mkdir -p gopath/src/github.com/karmarun/karma.run
RUN mkdir -p gopath/src/github.com/boltdb/bolt
RUN mkdir -p gopath/src/github.com/kr/pretty
RUN mkdir -p gopath/src/golang.org/x/crypto

# karmarun/karma.run tag 0.1.4-alpha
RUN git clone --branch 0.1.4-alpha --depth 1  \
    https://github.com/karmarun/karma.run.git \
    gopath/src/github.com/karmarun/karma.run

# boltdb/bolt tag v1.3.1
RUN git clone --branch v1.3.1 --depth 1 \
    https://github.com/boltdb/bolt.git  \
    gopath/src/github.com/boltdb/bolt

# kr/pretty branch master
RUN git clone --branch master --depth 1 \
    https://github.com/kr/pretty.git    \
    gopath/src/github.com/kr/pretty

# kr/text branch master
RUN git clone --branch master --depth 1 \
    https://github.com/kr/text.git    \
    gopath/src/github.com/kr/text

# x/crypto branch master
RUN git clone --branch master --depth 1  \
    https://github.com/golang/crypto.git \
    gopath/src/golang.org/x/crypto

WORKDIR gopath

RUN GOPATH=$(pwd) go build -o /karma.run github.com/karmarun/karma.run

WORKDIR /

RUN rm -rf /home/root/gopath

CMD "/karma.run"