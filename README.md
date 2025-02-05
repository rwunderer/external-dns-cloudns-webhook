# ExternalDNS - ClouDNS Webhook

⚠️ **This software is experimental.** ⚠️

[ExternalDNS](https://github.com/kubernetes-sigs/external-dns) is a Kubernetes
add-on for automatically DNS records for Kubernetes services using different
providers. By default, Kubernetes manages DNS records internally, but
ExternalDNS takes this functionality a step further by delegating the management
of DNS records to an external DNS provider such as this one. This webhook allows
you to manage your ClouDNS domains inside your kubernetes cluster.

## Requirements

[Auth-id and auth-password](https://www.cloudns.net/wiki/article/42/)
for the account managing your domains is required for this webhook to work
properly.

This webhook can be used in conjunction with **ExternalDNS v0.14.0 or higher**,
configured for using the webhook interface. Some examples for a working
configuration are shown in the next section.

## Kubernetes Deployment

The ClouDNS webhook is provided as a regular Open Container Initiative (OCI)
image released in the
[GitHub container registry](https://github.com/rwunderer/external-dns-cloudns-webhook/pkgs/container/external-dns-cloudns-webhook).
The deployment can be performed in every way Kubernetes supports.

Here are provided examples using the
[External DNS chart](#using-the-externaldns-chart) and the
[Bitnami chart](#using-the-bitnami-chart).

In either case, a secret that stores the CloudDNS auth info is required:

```yaml
kubectl create secret generic cloudns-config -n external-dns \
  --from-literal=CLOUDNS_AUTH_ID_TYPE=auth-id \
  --from-literal=CLOUDNS_AUTH_ID='<EXAMPLE_PLEASE_REPLACE>' \
  --from-literal=CLOUDNS_AUTH_PASSWORD='<EXAMPLE_PLEASE_REPLACE>'
```

### Using the ExternalDNS chart

Skip this step if you already have the ExternalDNS repository added:

```shell
helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
```

Update your helm chart repositories:

```shell
helm repo update
```

You can then create the helm values file, for example
`external-dns-cloudns-values.yaml`:

```yaml
namespace: external-dns
policy: sync
provider:
  name: webhook
  webhook:
    image:
      repository: ghcr.io/rwunderer/external-dns-cloudns-webhook
      tag: v0.1.0
    env:
    - name: CLOUDNS_AUTH_ID_TYPE
      valueFrom:
        secretKeyRef:
          name: cloudns-config
          key: CLOUDNS_AUTH_ID_TYPE
    - name: CLOUDNS_AUTH_ID
      valueFrom:
        secretKeyRef:
          name: cloudns-config
          key: CLOUDNS_AUTH_ID
    - name: CLOUDNS_AUTH_PASSWORD
      valueFrom:
        secretKeyRef:
          name: cloudns-config
          key: CLOUDNS_AUTH_PASSWORD
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      seccompProfile:
        type: RuntimeDefault
    livenessProbe:
      httpGet:
        path: /health
        port: http-webhook
      initialDelaySeconds: 10
      timeoutSeconds: 5
    readinessProbe:
      httpGet:
        path: /ready
        port: http-webhook
      initialDelaySeconds: 10
      timeoutSeconds: 5

extraArgs:
  - "--txt-prefix=reg-%{record_type}-"
```

And then:

```shell
# install external-dns with helm
helm install external-dns-cloudns external-dns/external-dns -f external-dns-cloudns-values.yaml --version 0.15.0 -n external-dns
```

### Using the Bitnami chart

Skip this step if you already have the Bitnami repository added:

```shell
helm repo add bitnami https://charts.bitnami.com/bitnami
```

Update your helm chart repositories:

```shell
helm repo update
```

You can then create the helm values file, for example
`external-dns-cloudns-values.yaml`:

```yaml
provider: webhook
policy: sync
extraArgs:
  webhook-provider-url: http://localhost:8888
  txt-prefix: "reg-%{record_type}-"

sidecars:
  - name: cloudns-webhook
    image: ghcr.io/rwunderer/external-dns-cloudns-webhook:v0.1.0
    ports:
      - containerPort: 8888
        name: webhook
      - containerPort: 8080
        name: http-wh-metrics
    livenessProbe:
      httpGet:
        path: /health
        port: http-wh-metrics
      initialDelaySeconds: 10
      timeoutSeconds: 5
    readinessProbe:
      httpGet:
        path: /ready
        port: http-wh-metrics
      initialDelaySeconds: 10
      timeoutSeconds: 5
    env:
    - name: CLOUDNS_AUTH_ID_TYPE
      valueFrom:
        secretKeyRef:
          name: cloudns-config
          key: CLOUDNS_AUTH_ID_TYPE
    - name: CLOUDNS_AUTH_ID
      valueFrom:
        secretKeyRef:
          name: cloudns-config
          key: CLOUDNS_AUTH_ID
    - name: CLOUDNS_AUTH_PASSWORD
      valueFrom:
        secretKeyRef:
          name: cloudns-config
          key: CLOUDNS_AUTH_PASSWORD
```

And then:

```shell
# install external-dns with helm
helm install external-dns-cloudns bitnami/external-dns -f external-dns-cloudns-values.yaml -n external-dns
```


## Environment variables

The following environment variables can be used for configuring the application.

### ClouDNS API calls configuration

These variables control the behavior of the webhook when interacting with
ClouDNS API.

| Variable              | Description                       | Notes                      |
| --------------------- | ----------------------------------| -------------------------- |
| CLOUDNS_AUTH_ID_TYPE  | either `auth-id` or `sub-auth-id` | Default: `auth-id`         |
| CLOUDNS_AUTH_ID       | ClouDNS auth-id or sub-auth-id    | Mandatory                  |
| CLOUDNS_AUTH_PASSWORD | ClouDNS auth-password             | Mandatory                  |
| DEFAULT_TTL           | Default record TTL                | Default: `3600`            |

### Test and debug

These environment variables are useful for testing and debugging purposes.

| Variable        | Description                      | Notes            |
| --------------- | -------------------------------- | ---------------- |
| DRY_RUN         | If set, changes won't be applied | Default: `false` |
| CLOUDNS_DEBUG   | Enables debugging messages       | Default: `false` |

### Socket configuration

These variables control the sockets that this application listens to.

| Variable        | Description                      | Notes                |
| --------------- | -------------------------------- | -------------------- |
| WEBHOOK_HOST    | Webhook hostname or IP address   | Default: `localhost` |
| WEBHOOK_PORT    | Webhook port                     | Default: `8888`      |
| METRICS_HOST    | Metrics hostname                 | Default: `0.0.0.0`   |
| METRICS_PORT    | Metrics port                     | Default: `8080`      |
| READ_TIMEOUT    | Sockets' read timeout in ms      | Default: `60000`     |
| WRITE_TIMEOUT   | Sockets' write timeout in ms     | Default: `60000`     |


### Domain filtering

Additional environment variables for domain filtering. When used, this webhook
will be able to work only on domains matching the filter.

| Environment variable           | Description                        |
| ------------------------------ | ---------------------------------- |
| DOMAIN_FILTER                  | Filtered domains                   |
| EXCLUDE_DOMAIN_FILTER          | Excluded domains                   |
| REGEXP_DOMAIN_FILTER           | Regex for filtered domains         |
| REGEXP_DOMAIN_FILTER_EXCLUSION | Regex for excluded domains         |

If the `REGEXP_DOMAIN_FILTER` is set, the following variables will be used to
build the filter:

 - REGEXP_DOMAIN_FILTER
 - REGEXP_DOMAIN_FILTER_EXCLUSION

 otherwise, the filter will be built using:

 - DOMAIN_FILTER
 - EXCLUDE_DOMAIN_FILTER

## Endpoints

This process exposes several endpoints, that will be available through these
sockets:

| Socket name | Socket address                |
| ----------- | ----------------------------- |
| Webhook     | `WEBHOOK_HOST`:`WEBHOOK_PORT` |
| Metrics     | `METRICS_HOST`:`METRICS_PORT` |

The environment variables controlling the socket addresses are not meant to be
changed, under normal circumstances, for the reasons explained in
[Tweaking the configuration](tweaking-the-configuration).
The endpoints
[expected by ExternalDNS](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/webhook-provider.md)
are marked with *.

### Webhook socket

All these endpoints are
[required by ExternalDNS](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/webhook-provider.md).

| Endpoint           | Purpose                                        |
| ------------------ | ---------------------------------------------- |
| `/`                | Initialization and `DomainFilter` negotiations |
| `/record`          | Get and apply records                          |
| `/adjustendpoints` | Adjust endpoints before submission             |

### Metrics socket

ExternalDNS doesn't have functional requirements for this endpoint, but some
of them are
[recommended](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/webhook-provider.md).
In this table those endpoints are marked with  __*__.

| Endpoint           | * | Purpose                                            |
| ------------------ | - | -------------------------------------------------- |
| `/health`          |   | Implements the liveness probe                      |
| `/ready`           |   | Implements the readiness probe                     |
| `/healthz`         | * | Implements a combined liveness and readiness probe |
| `/metrics`         | * | Exposes the available metrics                      |

Please check the [Exposed metrics](#exposed-metrics) section for more
information.

## Tweaking the configuration

While tweaking the configuration, there are some points to take into
consideration:

- if `WEBHOOK_HOST` and `METRICS_HOST` are set to the same address/hostname or
  one of them is set to `0.0.0.0` remember to use different ports. Please note
  that it **highly recommendend** for `WEBHOOK_HOST` to be `localhost`, as
  any address reachable from outside the pod might be a **security issue**;
  besides this, changing these would likely need more tweaks than just setting
  the environment variables. The default settings are compatible with the
  [ExternalDNS assumptions](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/webhook-provider.md);
- if your records don't get deleted when applications are uninstalled, you
  might want to verify the policy in use for ExternalDNS: if it's `upsert-only`
  no deletion will occur. It must be set to `sync` for deletions to be
  processed. Please check that `external-dns-cloudns-values.yaml` include:

  ```yaml
  policy: sync
  ```
- the `--txt-prefix` parameter should really include: `%{record_type}`, as any
  other value will cause a weird duplication of database records. Change the
  value provided in the sample configuration only if you really know what are
  you doing.

## Exposed metrics

The following metrics related to the API calls towards ClouDNS are available
for scraping.

| Name                         | Type      | Labels   | Description                                              |
| ---------------------------- | --------- | -------- | -------------------------------------------------------- |
| `successful_api_calls_total` | Counter   | `action` | The number of successful  API calls                      |
| `failed_api_calls_total`     | Counter   | `action` | The number of API calls that returned an error           |
| `filtered_out_zones`         | Gauge     | _none_   | The number of zones excluded by the domain filter        |
| `skipped_records`            | Gauge     | `zone`   | The number of skipped records per domain                 |
| `api_delay_hist`             | Histogram | `action` | Histogram of the delay (ms) when calling the ClouDNS API |

The label `action` can assume one of the following values, depending on the
ClouDNS API endpoint called:

- `get_zones`
- `get_records`
- `create_record`
- `delete_record`
- `update_record`

The label `zone` can assume one of the zone names as its value.

Please notice that in some cases an _update_ request from ExternalDNS will be
transformed into a `delete_record` and subsequent `create_record` calls by this
webhook.


## Development

The basic development tasks are provided by make. Run `make help` to see the
available targets.

## Credits

This Webhook was forked and modified from the [Hetzner Webhook](https://github.com/mconfalonieri/external-dns-hetzner-webhook)
to work with CloudDNS. The actual ClouDNS provider was taken from work by [wmarchesi123](https://github.com/wmarchesi123/external-dns/tree/cloudns).
