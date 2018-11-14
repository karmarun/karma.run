FROM alpine:latest AS build

RUN apk update && apk add git go libc-dev

RUN mkdir -p /gopath/src/karma.run

WORKDIR /gopath/src/karma.run
RUN git clone https://github.com/karmarun/karma.run.git .

WORKDIR /gopath
RUN GOPATH=$(pwd) go get karma.run
RUN mv bin/karma.run /karma.run
RUN rm -rf gopath

FROM alpine:latest AS deploy
COPY --from=build /karma.run /karma.run

WORKDIR /
CMD "/karma.run"
