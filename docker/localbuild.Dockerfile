FROM cgr.dev/chainguard/static@sha256:d4c20db9cb2dbf1ac9ec77f9dbc11080a78514a5f9b96096965550dbd1c73e09 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
