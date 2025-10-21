FROM cgr.dev/chainguard/static@sha256:bf076ce7861fe5cd50414b8ef26af247df58af0e256e17a7e4fc5ef2450393f9 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
