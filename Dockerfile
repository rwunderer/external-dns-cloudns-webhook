#--------
# builder
#--------
FROM golang:1.25.4-alpine@sha256:d2ede9f3341a67413127cf5366bb25bbad9b0a66e8173cae3a900ab00e84861f AS builder

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
FROM cgr.dev/chainguard/static@sha256:939a132511fcbc2702e0e251b6f3ea368c0ad4f114678ae5973903352357d01a AS external-dns-cloudns-webhook

LABEL version=0.3.31-rc0

USER 20000:20000

COPY --from=builder --chmod=555 /app/external-dns-cloudns-webhook /opt/external-dns-cloudns-webhook

ENTRYPOINT ["/opt/external-dns-cloudns-webhook"]
