FROM alpine

WORKDIR /root

COPY loadbalancer-controller /usr/bin/loadbalancer-controller

ENTRYPOINT ["/usr/bin/loadbalancer-controller"]

CMD ["--debug"]
