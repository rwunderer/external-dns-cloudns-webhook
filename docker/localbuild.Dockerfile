FROM cgr.dev/chainguard/static@sha256:e55a04f85f58e6b0e36bae05b8ff18c79035e65c0151e4e866eb49679782d28e AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
