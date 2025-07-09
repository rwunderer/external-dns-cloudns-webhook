FROM cgr.dev/chainguard/static@sha256:c9635595e59e9f4a48da16842ce8dd8984298af3140dcbe5ed2ea4a02156db9c AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
