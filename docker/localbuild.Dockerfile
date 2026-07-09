FROM cgr.dev/chainguard/static@sha256:60582b2ae6074f641094af0f370d4ab241aab271858a66223dcde7eee9f51638 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
