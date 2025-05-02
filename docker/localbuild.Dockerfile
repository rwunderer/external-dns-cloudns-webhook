FROM cgr.dev/chainguard/static@sha256:2a625816afa718bedc374daaed98b7171bb74591f10067f42efb448bfc8ea1ee AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
