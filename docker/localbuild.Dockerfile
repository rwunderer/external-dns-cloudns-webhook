FROM cgr.dev/chainguard/static@sha256:8d1a96321dca0e8e7848b7db2d431191f15e7e302faa1428100bbab351d42c7a AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
