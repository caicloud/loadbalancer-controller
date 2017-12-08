FROM cargo.caicloudprivatetest.com/caicloud/alpine:3.6

LABEL maintainer="Jun Zhang <jim.zhang@caicloud.io>"

WORKDIR /root

COPY bin/controller /usr/bin/controller

ENTRYPOINT ["/usr/bin/controller"]

CMD ["--debug"]
