FROM cgr.dev/chainguard/static@sha256:11ec91f0372630a2ca3764cea6325bebb0189a514084463cbb3724e5bb350d14 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
