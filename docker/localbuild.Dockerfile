FROM cgr.dev/chainguard/static@sha256:b24ac9892a647b64bc8341e6b9280885c44004b32ffc8d57d3218960280e1d89 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
