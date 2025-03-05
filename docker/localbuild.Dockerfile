FROM cgr.dev/chainguard/static@sha256:7a6456cc96ecde793b7c8ad9a3ccd5d610d6168a6f64d693ecc2e84f8276c6c6 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
