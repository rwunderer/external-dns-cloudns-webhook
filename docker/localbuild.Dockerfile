FROM cgr.dev/chainguard/static@sha256:324c96273762d9500fd72d973f7d05f0dd15be0668935b3ba02221658041dc9a AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
