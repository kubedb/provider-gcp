# AGENTS.md - provider-gcp

This file provides instructions for AI coding agents working in the KubeDB `provider-gcp` repository.

## Project Overview

`provider-gcp` is a [Crossplane](https://crossplane.io/) provider for Google Cloud Platform, generated using the [Upjet](https://github.com/upbound/upjet) code-generation pipeline on top of the upstream [`hashicorp/terraform-provider-google`](https://github.com/hashicorp/terraform-provider-google) Terraform provider. It exposes XRM-conformant Kubernetes managed resources for selected GCP services (Compute networking, Cloud SQL, Spanner, Memorystore Redis) under the `gcp.kubedb.com` API root group. Module path: `kubedb.dev/provider-gcp` (Go 1.22).

The provider ships two binaries: `provider` (the controller manager) and `generator` (the code-generation entrypoint). Generated CRD/types/controller code is committed to the repo (files prefixed with `zz_`).

## Build & Development Commands

```bash
# Run code-generation pipeline (regenerates apis/, internal/controller/, package/crds/)
go run cmd/generator/main.go "$PWD"

# Build provider binaries (provider, generator)
make build

# Run the provider locally, out-of-cluster (uses current kubeconfig)
make run

# Build, push images, and install
make all

# Generate Terraform provider schema (config/schema.json)
make config/schema.json

# Pull upstream Terraform provider docs into .work/
make pull-docs

# Generate cobertura coverage report (excludes zz_ generated files)
make cobertura

# Update git submodules (build/ submodule with shared makelib)
make submodules
```

### End-to-end testing

```bash
# Build, install on a controlplane, and run uptest-based e2e
make e2e

# Run uptest only against an already-deployed provider
make uptest UPTEST_EXAMPLE_LIST=examples/redis/provision-redis-instance.yaml
```

### Terraform versions pinned in Makefile

| Variable | Value |
|----------|-------|
| `TERRAFORM_VERSION` | `1.3.3` |
| `TERRAFORM_PROVIDER_SOURCE` | `hashicorp/google` |
| `TERRAFORM_PROVIDER_VERSION` | `4.69.1` |
| `TERRAFORM_NATIVE_PROVIDER_BINARY` | `terraform-provider-google_v4.69.1_x5` |
| `GO_REQUIRED_VERSION` | `1.25` |
| `GOLANGCILINT_VERSION` | `1.50.0` |

## Project Structure

```
apis/
  compute/v1alpha1/        # Generated CRD Go types: Firewall, Network, NetworkPeering
  sql/v1alpha1/            # Cloud SQL: Database, DatabaseInstance, User, SSLCert, SourceRepresentationInstance
  spanner/v1alpha1/        # Spanner: Instance, Database, InstanceIAMMember, DatabaseIAMMember
  redis/v1alpha1/          # Memorystore: Instance
  v1alpha1/                # StoreConfig (alpha external secret stores)
  v1beta1/                 # ProviderConfig, ProviderConfigUsage
  register_crd.go          # AddToSchemeCrd (apiextensions.k8s.io types)
  zz_register.go           # Aggregate AddToScheme builder (generated)
cmd/
  provider/main.go         # Controller-manager entrypoint (kingpin flags)
  generator/main.go        # Upjet pipeline driver: pipeline.Run + dynamic-controller gen
  dynamic-controller/      # Dynamic CRD reconciler generator (setup.go, crd_controller.go.txt template)
config/
  provider.go              # Upjet provider config (root group gcp.kubedb.com, resource prefix gcp)
  external_name.go         # External-name templates per Terraform resource
  overrides.go             # Group/kind overrides
  schema.json              # Embedded Terraform provider schema (//go:embed)
  provider-metadata.yaml   # Embedded Terraform provider metadata
  compute/, sql/, spanner/, redis/  # Per-service Upjet Configure() functions
  common/                  # Shared config helpers
internal/
  clients/gcp.go           # TerraformSetupBuilder: builds terraform.Setup from ProviderConfig + credentials secret
  controller/              # Generated controllers per resource + zz_setup.go and zz_dynamic_crd_controller.go
  features/features.go     # Feature flags: EnableAlphaExternalSecretStores, EnableBetaManagementPolicies
package/
  crds/                    # Generated CRD manifests (one YAML per kind)
  crossplane.yaml          # Crossplane package metadata
examples/                  # Hand-written CR examples per service
examples-generated/        # Generated CR examples
cluster/test/setup.sh      # Bootstrap script invoked by `make uptest`
build/                     # Git submodule (build/makelib/*.mk shared Crossplane build system)
hack/boilerplate.go.txt    # License header injected by upjet generators
```

## Key Packages / APIs

- `kubedb.dev/provider-gcp/config` (`GetProvider`): builds the Upjet `*ujconfig.Provider` from `schema.json` and `provider-metadata.yaml`, sets root group `gcp.kubedb.com`, applies `ExternalNameConfigurations`, and dispatches per-service `Configure` funcs (`compute`, `sql`, `spanner`, `redis`).
- `kubedb.dev/provider-gcp/internal/clients` (`TerraformSetupBuilder`): resolves a referenced `ProviderConfig`, extracts credentials JSON via `resource.CommonCredentialExtractor`, tracks usage, and returns a `terraform.Setup` populated with `project` and `credentials`.
- `kubedb.dev/provider-gcp/internal/controller` (`zz_setup.go`, `NewCustomResourceReconciler`): wires per-resource Upjet reconcilers into the controller-runtime manager. `zz_dynamic_crd_controller.go` is generated by `cmd/dynamic-controller`.
- `kubedb.dev/provider-gcp/apis/v1beta1`: `ProviderConfig` (with `spec.projectID` and `spec.credentials.source` selectors) and `ProviderConfigUsage`.
- `kubedb.dev/provider-gcp/apis/v1alpha1`: `StoreConfig` (alpha external secret stores, gated by `--enable-external-secret-stores`).

### External-name conventions (config/external_name.go)

Most resources use `config.TemplatedStringAsIdentifier` keyed off `.setup.configuration.project` to produce the Terraform import ID:

| Terraform resource | External-name template |
|---|---|
| `google_compute_network` | `projects/{project}/global/networks/{external_name}` |
| `google_compute_firewall` | `projects/{project}/global/firewalls/{external_name}` |
| `google_sql_database_instance` | `projects/{project}/instances/{external_name}` |
| `google_sql_database` | `projects/{project}/instances/{instance}/databases/{external_name}` |
| `google_sql_user` | `{project}/{instance}/{external_name}` |
| `google_spanner_instance` | `{project}/instance/{external_name}` |
| `google_redis_instance` | `projects/{project}/locations/{region}/instances/{external_name}` |
| `google_sql_ssl_cert`, `google_spanner_*_iam_member` | `IdentifierFromProvider` |

## Provider runtime flags (cmd/provider/main.go)

```
--debug                          (-d)  Run with debug logging
--sync                           (-s)  Controller manager sync period (default 1h)
--leader-election                (-l)  Use leader election (env LEADER_ELECTION)
--terraform-version                    (required, env TERRAFORM_VERSION)
--terraform-provider-source            (required, env TERRAFORM_PROVIDER_SOURCE)
--terraform-provider-version           (required, env TERRAFORM_PROVIDER_VERSION)
--max-reconcile-rate                   Global QPS for drift checks (default 10)
--namespace                            Default scope for external secret stores (env POD_NAMESPACE)
--enable-external-secret-stores        Alpha (env ENABLE_EXTERNAL_SECRET_STORES)
--enable-management-policies           Beta  (env ENABLE_MANAGEMENT_POLICIES)
```

`main.go` registers both `apis.AddToScheme` and `apis.AddToSchemeCrd`, then starts `NewCustomResourceReconciler` (dynamic CRD reconciler) rather than the static `controller.Setup`.

## Testing

There are no checked-in unit tests for resource controllers; validation is via:

1. `make lint` / `make test` (delegated to `build/makelib/golang.mk`).
2. `make e2e` - builds the provider, deploys it to a Crossplane controlplane via `controlplane.up` + `local.xpkg.deploy`, then runs `uptest` (`UPTEST_VERSION=v0.2.1`) against `examples/`.
3. `cobertura` target excludes generated `zz_*` files from coverage.

CI workflows live in `.github/workflows/` (`ci.yml`, `e2e.yml`, `release.yml`, `release-tracker.yml`, `backport.yml`, `tag.yml`, `commands.yml`). CI pins `GO_VERSION=1.20` and `GOLANGCI_VERSION=v1.53.3` (note: differs from Makefile defaults).

## Dependencies

Selected direct deps from `go.mod` (Go 1.22, toolchain go1.22.2):

| Module | Version |
|---|---|
| `github.com/crossplane/crossplane-runtime` | `v1.15.1` |
| `github.com/crossplane/crossplane-tools` | `v0.0.0-20230925130601-628280f8bf79` |
| `github.com/crossplane/upjet` | `v1.0.0` |
| `github.com/muvaf/typewriter` | `v0.0.0-20220131201631-921e94e8e8d7` |
| `sigs.k8s.io/controller-runtime` | `v0.17.2` |
| `sigs.k8s.io/controller-tools` | `v0.14.0` |
| `k8s.io/apimachinery`, `client-go`, `apiextensions-apiserver` | `v0.29.2` |
| `gopkg.in/alecthomas/kingpin.v2` | `v2.2.6` |

Indirect: HashiCorp Terraform plugin SDK v2 (`v2.24.0`), `terraform-plugin-go v0.14.0`, `hcl/v2 v2.14.1`.

## Code Conventions

- Files prefixed with `zz_` are generated by Upjet / controller-tools / typewriter - **do not edit by hand**. Regenerate via `go run cmd/generator/main.go "$PWD"`.
- License headers come from `hack/boilerplate.go.txt`.
- Resource API groups follow `<service>.gcp.kubedb.com` (e.g., `sql.gcp.kubedb.com`, `compute.gcp.kubedb.com`); the root group `gcp.kubedb.com` is used for `ProviderConfig` and `StoreConfig`.
- Resource prefix in Terraform source: `gcp` (e.g., `google_sql_database` -> stripped to `sql_database`); group/kind is determined by `groupKindOverride` in `config/provider.go`.
- New resource bring-up: (1) add an entry to `ExternalNameConfigs` in `config/external_name.go`, (2) add the resource name to `ExternalNameConfigured()` include list, (3) add per-service `Configure()` in `config/<service>/`, (4) regenerate.
- Examples live in `examples/<service>/`; generated examples in `examples-generated/` should not be committed by hand.
- `.golangci.yml` defines lint rules; `make reviewable` (from shared makelib) runs full pre-PR checks.
- Linux build platforms: `linux_amd64 linux_arm64` (`PLATFORMS` in Makefile).
- Image registry: `ghcr.io/kubedb` (also published to `xpkg.upbound.io/upbound`).
