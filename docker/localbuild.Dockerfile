FROM cgr.dev/chainguard/static@sha256:797e62f43d04d792e9f930913e7d9f5a63e92bd19ca5e7e5139a692decee2dbc AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
