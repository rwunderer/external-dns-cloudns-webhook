#--------
# builder
#--------
FROM golang:1.25.1-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd AS builder

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
FROM cgr.dev/chainguard/static@sha256:b2e1c3d3627093e54f6805823e73edd17ab93d6c7202e672988080c863e0412b AS external-dns-cloudns-webhook

LABEL version=0.3.27-rc0

USER 20000:20000

COPY --from=builder --chmod=555 /app/external-dns-cloudns-webhook /opt/external-dns-cloudns-webhook

ENTRYPOINT ["/opt/external-dns-cloudns-webhook"]
