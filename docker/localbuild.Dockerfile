FROM cgr.dev/chainguard/static@sha256:1ff7590cbc50eaaa917c34b092de0720d307f67d6d795e4f749a0b80a2e95a2c AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
