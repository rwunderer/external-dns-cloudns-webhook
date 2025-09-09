FROM cgr.dev/chainguard/static@sha256:7124bf9a6f70e0750d14ef16f1791f322f6d62f50a49223a709f7ed41644c353 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
