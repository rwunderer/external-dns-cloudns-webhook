FROM cgr.dev/chainguard/static@sha256:2e3db1641bb4fe4e85d2210f4aadb79252e90d5fa745f53a3ffed6a1aab4f73b AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
