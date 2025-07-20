FROM cgr.dev/chainguard/static@sha256:93b70336be10c325d5a96016971b71b38d8e79e5148af2144f2aae93ee9367c3 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
