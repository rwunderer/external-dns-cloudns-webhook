FROM cgr.dev/chainguard/static@sha256:7d8e6efa03a7b58b5a5b2a1d8555e44b990775b29d6324e12d1c77314d595aaa AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
