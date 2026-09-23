FROM cgr.dev/chainguard/static@sha256:a4e031f6d32f1af65a85b6925a9489db7787f09c73eefcd81793215c904d515d AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
