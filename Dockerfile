FROM golang:1.17 AS build-go

COPY . /src

RUN cd /src/ && go build -x

FROM ubuntu:latest

COPY --from=build-go /src/crossjoin /bin/crossjoin

WORKDIR /crossjoin

ENTRYPOINT ["/bin/crossjoin"]
