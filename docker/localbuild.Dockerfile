FROM cgr.dev/chainguard/static@sha256:0f33d0da1c868b246fffb2b951754689bdb5de60771157ed0a6a149a9be856f6 AS external-dns-cloudns-webhook
ARG TARGETARCH
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-$TARGETARCH /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
