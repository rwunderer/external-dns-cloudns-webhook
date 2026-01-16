FROM cgr.dev/chainguard/static@sha256:3348c5f7b97a4d63944034a8c6c43ad8bc69771b2564bed32ea3173bc96b4e04 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
