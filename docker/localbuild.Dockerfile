FROM cgr.dev/chainguard/static@sha256:9e2306b457ff71ce370043dc68c7b52bb1e6622c1c6c6714edbed57496f3116e AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
