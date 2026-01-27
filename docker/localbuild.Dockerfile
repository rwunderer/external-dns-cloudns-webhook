FROM cgr.dev/chainguard/static@sha256:99a5f826e71115aef9f63368120a6aa518323e052297718e9bf084fb84def93c AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
