FROM cgr.dev/chainguard/static@sha256:3c0cfe403ce6f1ff7761aab482448e5d4979762a9853103639c0d769aa7de89e AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
