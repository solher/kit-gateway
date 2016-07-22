FROM golang:1.6-alpine

ENV APP_NAME="kit-gateway"
ENV SRC_PATH="/go/src/github.com/solher/kit-gateway"

RUN apk add --update git \
&& mkdir -p $SRC_PATH
COPY . $SRC_PATH
WORKDIR $SRC_PATH

RUN go get -u ./... \
&& go build -v \
&& cp $APP_NAME /usr/bin \
&& apk del git \
&& rm -rf /go/* \
&& adduser -D app

WORKDIR /

USER app
EXPOSE 3000
CMD $APP_NAME -crud.addr="kit-crud:8082" -zipkin.addr="zipkin:9410"