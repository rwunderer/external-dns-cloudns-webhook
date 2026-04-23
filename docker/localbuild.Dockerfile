FROM cgr.dev/chainguard/static@sha256:bee469c98ce2df388a2746dc360bd59eb3efe2dab366a01cdcbfd738a2ca1474 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
