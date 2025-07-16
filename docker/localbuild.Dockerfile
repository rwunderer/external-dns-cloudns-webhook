FROM cgr.dev/chainguard/static@sha256:d786d1c686ce4a49376cd2f068d91e691b2bb2e3a6f38513b2396b69b1a9c06f AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
