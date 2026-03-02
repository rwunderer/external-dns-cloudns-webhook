FROM cgr.dev/chainguard/static@sha256:d6a97eb401cbc7c6d48be76ad81d7899b94303580859d396b52b67bc84ea7345 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
