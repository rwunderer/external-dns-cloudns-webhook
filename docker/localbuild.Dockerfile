FROM cgr.dev/chainguard/static@sha256:2fdfacc8d61164aa9e20909dceec7cc28b9feb66580e8e1a65b9f2443c53b61b AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
