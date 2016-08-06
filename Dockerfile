FROM golang:1.6-alpine

ENV APP_NAME="kit-gateway"
ENV SRC_PATH="/go/src/github.com/solher/kit-gateway"

ADD https://raw.githubusercontent.com/solher/env2flags/master/env2flags.sh /usr/local/bin/env2flags
RUN chmod u+x /usr/local/bin/env2flags

RUN apk add --update git \
&& mkdir -p $SRC_PATH
COPY . $SRC_PATH
WORKDIR $SRC_PATH

RUN go get -u ./... \
&& go build -v \
&& cp $APP_NAME /usr/local/bin \
&& apk del git \
&& rm -rf /go/*

EXPOSE 3000
CMD ["$APP_NAME"]