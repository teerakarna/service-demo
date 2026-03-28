# service-demo

Demo Items CRUD API — Go + Gin, distroless container, Helm chart.
Part of the [From Commit to Cluster](https://teerakarna.github.io/posts/gitops-pipeline-demo/) tutorial.

## API

```
GET    /healthz
GET    /readyz
GET    /api/v1/items
POST   /api/v1/items        {"name": "...", "description": "..."}
GET    /api/v1/items/:id
PUT    /api/v1/items/:id    {"name": "...", "description": "..."}
DELETE /api/v1/items/:id
```

## Run locally

```bash
go run ./cmd/server
curl http://localhost:8080/healthz
```

## Test

```bash
# Unit tests
go test ./internal/... -v -race

# Integration tests
go test ./integration/... -v

# All
go test ./... -v
```

## Lint & SCA

```bash
golangci-lint run
govulncheck ./...
```

## Build image

```bash
docker build -t service-demo:local .
docker run -p 8080:8080 service-demo:local
```

## Pipeline

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` | PR opened / updated | Lint, SCA, unit tests, integration tests, build image, deploy ephemeral env, smoke tests |
| `cd.yml` | Push to main | Build `main-{sha}` image, promote to dev + preprod, run preprod tests |
| `release.yml` | Manual (`workflow_dispatch`) | Verify preprod gate, retag image to `v{X.Y.Z}`, promote to prod |

## Required secrets

| Secret | Used by | Description |
|---|---|---|
| `GITHUB_TOKEN` | All | Auto-provided by Actions — GHCR push |
| `GITOPS_TOKEN` | `cd.yml`, `release.yml` | PAT with write access to `gitops-demo` |
| `ARGOCD_TOKEN` | `ci.yml`, `cd.yml` | ArgoCD API token for `argocd app wait` |

## Self-hosted runner

The `deploy-ephemeral`, `smoke-test`, and `preprod-tests` jobs require a runner with
access to the local kind cluster. Register one with the label `kind`:

```bash
# In service-demo repo: Settings → Actions → Runners → New self-hosted runner
# Then on your machine, after downloading the runner:
./run.sh --labels kind
```

## Postman

Import `postman/service-demo.postman_collection.json`. Set the `base_url` variable to
your target environment. Port-forward first:

```bash
kubectl port-forward -n <env> svc/service-demo 8080:80
```

Run the collection in order — Create Item populates `item_id` for subsequent requests.
