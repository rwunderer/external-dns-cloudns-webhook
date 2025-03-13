FROM cgr.dev/chainguard/static@sha256:beb6a9eaf915a03a6fedbeda117fd327cd6b08883ae5fa58bd2ac7c0980318cd AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
