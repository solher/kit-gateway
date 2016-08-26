FROM alpine:latest

ADD https://raw.githubusercontent.com/solher/env2flags/master/env2flags.sh /usr/local/bin/env2flags
RUN chmod u+x /usr/local/bin/env2flags

COPY ./kit-gateway /usr/local/bin

WORKDIR /

EXPOSE 3000
# ENTRYPOINT ["env2flags", "APPDASH_ADDR", "CRUD_ADDR", "--"]
CMD ["kit-gateway"]