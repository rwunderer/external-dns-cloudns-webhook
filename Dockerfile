#--------
# builder
#--------
FROM golang:1.25.6-alpine@sha256:660f0b83cf50091e3777e4730ccc0e63e83fea2c420c872af5c60cb357dcafb2 AS builder

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
FROM cgr.dev/chainguard/static@sha256:99a5f826e71115aef9f63368120a6aa518323e052297718e9bf084fb84def93c AS external-dns-cloudns-webhook

LABEL version=0.4.2

USER 20000:20000

COPY --from=builder --chmod=555 /app/external-dns-cloudns-webhook /opt/external-dns-cloudns-webhook

ENTRYPOINT ["/opt/external-dns-cloudns-webhook"]
