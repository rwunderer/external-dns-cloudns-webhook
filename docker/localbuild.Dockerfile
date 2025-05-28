FROM cgr.dev/chainguard/static@sha256:633aabd19a2d1b9d4ccc1f4b704eb5e9d34ce6ad231a4f5b7f7a3af1307fdba8 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
