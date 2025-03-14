FROM cgr.dev/chainguard/static@sha256:9276a4ebe6b98cd1bbd53b8139228434a0e4f00d06d39e33688e9bd759986656 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
