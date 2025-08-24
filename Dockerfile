#--------
# builder
#--------
FROM golang:1.25.0-alpine@sha256:f18a072054848d87a8077455f0ac8a25886f2397f88bfdd222d6fafbb5bba440 AS builder

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
FROM cgr.dev/chainguard/static@sha256:6a4b683f4708f1f167ba218e31fcac0b7515d94c33c3acf223c36d5c6acd3783 AS external-dns-cloudns-webhook

LABEL version=0.3.21-rc0

USER 20000:20000

COPY --from=builder --chmod=555 /app/external-dns-cloudns-webhook /opt/external-dns-cloudns-webhook

ENTRYPOINT ["/opt/external-dns-cloudns-webhook"]
