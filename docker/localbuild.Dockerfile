FROM cgr.dev/chainguard/static@sha256:96d02f455d5a73b817c0602910748609cf8471b1cc9522f78c75cedb1f67d072 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
