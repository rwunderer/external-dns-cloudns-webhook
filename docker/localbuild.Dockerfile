FROM cgr.dev/chainguard/static@sha256:81b61e16687f76ebc3c1fa71ec3fa3e0901e2908e0cd442f378557c294920aac AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
