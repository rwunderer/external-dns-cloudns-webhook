FROM cgr.dev/chainguard/static@sha256:95a45fc5fda9aa71dbdc645b20c6fb03f33aec8c1c2581ef7362b1e6e1d09dfb AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
