FROM cgr.dev/chainguard/static@sha256:9047a7c9e2d7dbc6614053e82c859e0c409591ccf054447fdc375a5e5741f70b AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
