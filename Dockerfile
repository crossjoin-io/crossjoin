FROM node AS build-ui

COPY . /src

RUN cd /src/ui && npm install && npm run build

FROM golang:1.17 AS build-go

COPY . /src

COPY --from=build-ui /src/ui/public/js/app.js /src/ui/public/js/app.js
RUN cd /src/ && go build -x -ldflags "-linkmode external -extldflags -static"

FROM alpine:3

COPY --from=build-go /src/crossjoin /bin/crossjoin

WORKDIR /crossjoin

ENTRYPOINT ["/bin/crossjoin"]
