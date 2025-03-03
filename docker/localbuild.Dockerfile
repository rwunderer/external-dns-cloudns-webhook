FROM cgr.dev/chainguard/static@sha256:3e9af2550ae5ff1fe5b9d69332955c01213c37c75874b184e5fbea500d1c9808 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
