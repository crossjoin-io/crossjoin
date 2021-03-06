FROM node AS build-ui

COPY . /src

RUN cd /src/ui && npm install && npm run build

FROM golang:1.17 AS build-go

COPY . /src

COPY --from=build-ui /src/ui/public/js/app.js /src/ui/public/js/app.js
COPY --from=build-ui /src/ui/public/js/app.css /src/ui/public/js/app.css
RUN cd /src/ && go build -x -ldflags "-linkmode external -extldflags -static"

FROM alpine:3

RUN apk add docker

COPY --from=build-go /src/crossjoin /bin/crossjoin

WORKDIR /crossjoin

ENTRYPOINT ["/bin/crossjoin"]
