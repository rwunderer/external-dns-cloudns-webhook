FROM --platform=arm64 gcr.io/distroless/static-debian11:nonroot@sha256:63ebe035fbdd056ed682e6a87b286d07d3f05f12cb46f26b2b44fc10fc4a59ed
USER 20000:20000
ADD --chmod=555 build/bin/external-dns-cloudns-webhook-arm64 /opt/external-dns-cloudns-webhook/app

ENTRYPOINT ["/opt/external-dns-cloudns-webhook/app"]
