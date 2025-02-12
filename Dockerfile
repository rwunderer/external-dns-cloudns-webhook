#--------
# builder
#--------
FROM golang:1.24.0-alpine@sha256:5429efb7de864db15bd99b91b67608d52f97945837c7f6f7d1b779f9bfe46281 AS builder

ARG TARGETPLATFORM
ARG TARGETOS="linux"
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#"v"} go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o ./external-dns-cloudns-webhook ./cmd/webhook

#--------
# container
#--------
FROM cgr.dev/chainguard/static@sha256:853bfd4495abb4b65ede8fc9332513ca2626235589c2cef59b4fce5082d0836d AS external-dns-cloudns-webhook

LABEL version=0.2.2

USER 20000:20000

COPY --from=builder --chmod=555 /app/external-dns-cloudns-webhook /opt/external-dns-cloudns-webhook

ENTRYPOINT ["/opt/external-dns-cloudns-webhook"]
