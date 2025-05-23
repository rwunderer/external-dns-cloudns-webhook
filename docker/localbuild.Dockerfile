FROM cgr.dev/chainguard/static@sha256:48278935856fba0e9fac80365ae9a5b33297f7e5682c2dcb86ecfe5eb6878972 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
