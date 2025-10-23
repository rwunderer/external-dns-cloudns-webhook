FROM cgr.dev/chainguard/static@sha256:01f45a2a6b87a54e242361c217335b4e792b09b92cd4b0780f8b253e27d299bb AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
