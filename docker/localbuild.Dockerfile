FROM cgr.dev/chainguard/static@sha256:939a132511fcbc2702e0e251b6f3ea368c0ad4f114678ae5973903352357d01a AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
