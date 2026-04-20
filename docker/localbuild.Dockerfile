FROM cgr.dev/chainguard/static@sha256:1f14279403150757d801f6308bb0f4b816b162fddce10b9bd342f10adc3cf7fa AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
