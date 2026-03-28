# service-demo

Demo Items CRUD API — Go + Gin, distroless container, Helm chart.
Part of the [From Commit to Cluster](https://teerakarna.github.io/posts/gitops-pipeline-demo/) tutorial.

This repo contains the application code, Helm chart, and GitHub Actions workflows. The GitOps configuration (ArgoCD Applications, per-environment values) lives in [`gitops-demo`](https://github.com/teerakarna/gitops-demo). The CI/CD pipelines commit to that repo; ArgoCD reconciles the cluster from it.

---

## Repository structure

```
cmd/server/               # Entrypoint — reads PORT and ENV_NAME env vars
internal/
  api/                    # Gin router and HTTP handlers
  store/                  # Store interface + thread-safe in-memory implementation
integration/              # Full lifecycle integration tests (httptest.Server)
chart/service-demo/       # Helm chart — deployed by ArgoCD via gitops-demo
  values.yaml             # Defaults; overridden per environment in gitops-demo
  templates/
    deployment.yaml       # Exposes ENV_NAME and SERVICES_NAMESPACE env vars
    service.yaml
.github/workflows/
  ci.yml                  # PR: lint, SCA, tests, build, ephemeral env, smoke tests
  cd.yml                  # Merge to main: build main-{sha}, promote to preprod
  release.yml             # Manual: preflight gate, retag, promote prod + dev
postman/                  # Postman collection for manual API testing
Dockerfile                # Multi-stage: golang:1.22-alpine builder → distroless nonroot
```

---

## API

```
GET    /healthz                     liveness probe
GET    /readyz                      readiness probe
GET    /api/v1/items                list all items
POST   /api/v1/items                create an item  {"name": "...", "description": "..."}
GET    /api/v1/items/:id            get one item
PUT    /api/v1/items/:id            update an item  {"name": "...", "description": "..."}
DELETE /api/v1/items/:id            delete an item
```

The store is in-memory — no database, no external dependencies. The service is a single static binary.

---

## Run locally

```bash
go run ./cmd/server
curl http://localhost:8080/healthz
```

Override the port with `PORT=9090 go run ./cmd/server`.

---

## Test

```bash
# Unit tests (with race detector)
go test ./internal/... -v -race

# Integration tests
go test ./integration/... -v

# All
go test ./... -v
```

---

## Lint & SCA

```bash
# Static analysis
golangci-lint run

# Dependency vulnerability scan (Go vuln DB)
govulncheck ./...
```

Both run on every PR in CI. `govulncheck` fails on any vulnerability with a known fix that affects reachable code.

---

## Container image

```bash
docker build -t service-demo:local .
docker run -p 8080:8080 service-demo:local
```

The image uses a multi-stage build: `golang:1.22-alpine` compiles a statically linked binary; `gcr.io/distroless/static:nonroot` is the runtime base. No shell, no package manager, ~12 MB final image. The container runs as uid 65532 (nonroot) with `readOnlyRootFilesystem` and all capabilities dropped.

---

## Pipeline

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` | PR opened / updated | Lint, SCA, unit tests, integration tests, build `pr-{N}-{sha}` image, Trivy scan, deploy ephemeral env `pr-{N}`, smoke tests |
| `cd.yml` | Merge to main | Build `main-{sha}` image, Trivy scan, promote to preprod, run preprod integration tests, post check run |
| `release.yml` | Manual (`workflow_dispatch`) | Verify preprod gate, retag `main-{sha}` → `v{X.Y.Z}` (no rebuild), promote prod + dev values atomically, create GitHub Release |

### Job dependency graph

**CI:**
```
lint-sca ──┐
           ├──→ build ──→ deploy-ephemeral ──→ smoke-test
test    ───┘
```

**CD:**
```
build ──→ promote-preprod ──→ preprod-tests
```

**Release:**
```
preflight ──→ tag-and-promote
```

### Image tag scheme

| Tag format | Created by | Deployed to |
|---|---|---|
| `pr-{N}-{sha}` | `ci.yml` on PR push | Ephemeral `pr-{N}` namespace |
| `main-{sha}` | `cd.yml` on merge to main | `preprod` namespace |
| `v{X.Y.Z}` | `release.yml` retag (no rebuild) | `prod` and `dev` namespaces |

The retag at release is a pointer operation — `v{X.Y.Z}` and `main-{sha}` point to the same image digest. The bits that ran in preprod are byte-for-byte identical to what goes to production.

### The preprod gate

`cd.yml` posts a `CD / Preprod Tests` check run against the merge commit SHA. `release.yml` checks this before proceeding — if preprod tests are not passing, the release is blocked. There is no override.

---

## Required secrets

| Secret | Used by | Description |
|---|---|---|
| `GITHUB_TOKEN` | All | Auto-provided by Actions — GHCR image push |
| `GITOPS_TOKEN` | `ci.yml`, `cd.yml`, `release.yml` | PAT with write access to `gitops-demo` |
| `ARGOCD_TOKEN` | `ci.yml`, `cd.yml` | ArgoCD API token for `argocd app wait` |

---

## Self-hosted runner

The `deploy-ephemeral`, `smoke-test`, and `preprod-tests` jobs run on `[self-hosted, kind]` — they need access to the local kind cluster, which GitHub-hosted runners cannot reach.

Register a runner with the `kind` label in **Settings → Actions → Runners → New self-hosted runner**, then:

```bash
# After downloading and configuring the runner:
./run.sh --labels kind
```

The runner needs `kubectl` configured for the kind cluster and the `argocd` CLI installed.

---

## Postman

Import `postman/service-demo.postman_collection.json`. Set the `base_url` collection variable to your target environment. Port-forward first:

```bash
kubectl port-forward -n <env> svc/service-demo 8080:80
```

Run the collection in order — Create Item populates `item_id` automatically for subsequent requests.
