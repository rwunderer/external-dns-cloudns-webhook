FROM cgr.dev/chainguard/static@sha256:6d508f497fe786ba47d57f4a3cffce12ca05c04e94712ab0356b94a93c4b457f AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
