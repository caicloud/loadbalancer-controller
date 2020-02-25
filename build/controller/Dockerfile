FROM cargo.caicloud.xyz/library/debian:stretch

LABEL maintainer="Jun Zhang <jim.zhang@caicloud.io>"

WORKDIR /root

COPY bin/controller /usr/bin/controller

ENTRYPOINT ["/usr/bin/controller"]

CMD ["--v=2"]
