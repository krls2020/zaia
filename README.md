# ZAIA — Go CLI for AI Agents on Zerops

ZAIA is a standalone Go CLI binary designed for AI agents working with [Zerops](https://zerops.io) PaaS. It provides structured JSON output, BM25 knowledge search, YAML validation, and full service lifecycle management.

ZAIA is part of the **Zerops Control Plane (ZCP)** — a three-component system:

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│  ZAIA-MCP       │────>│  ZAIA CLI    │     │  ZCLI       │
│  (MCP server)   │     │  (this repo) │     │  (production)│
│  thin wrapper   │────>│  business    │     │  deploy,push │
└─────────────────┘     │  logic       │     └─────────────┘
                        └──────────────┘
                              │
                              v
                     ┌─────────────────┐
                     │ Zerops Cloud API│
                     └─────────────────┘
```

| Component | Role |
|-----------|------|
| **ZAIA CLI** (this repo) | All business logic: auth, discovery, knowledge, validation |
| **[ZAIA-MCP](https://github.com/krls2020/zaia-mcp)** | Thin MCP layer — calls `zaia` as subprocess |
| **ZCLI** | Production CLI — ZAIA does not depend on it |

## Features

- **Authentication** — `zaia login --token <value>` with auto project discovery (token must have access to exactly 1 project)
- **Service Discovery** — List services, env vars, service details
- **Logs** — Fetch service logs with severity/time filtering
- **Service Management** — Start, stop, restart, scale (async, returns process IDs)
- **Environment Variables** — Get/set/delete for both service and project scope
- **Import** — Import services from YAML (without `project:` section)
- **Delete** — Delete services (not projects) with confirmation
- **Subdomain** — Enable/disable Zerops subdomains (idempotent)
- **Validation** — Offline YAML validation for zerops.yml and import.yml
- **Knowledge Search** — BM25 full-text search over 65 embedded Zerops docs
- **Process Tracking** — Query and cancel async operations

## CLI Commands

### Auth & Info

| Command | Type | Description |
|---------|------|-------------|
| `zaia login --token <value>` | sync | Auto project discovery, stores to zaia.data |
| `zaia login --token <value> --url api.staging.zerops.io` | sync | Login to staging |
| `zaia logout` | sync | Removes zaia.data |
| `zaia status` | sync | Shows user + project (no API call) |
| `zaia version` | sync | Version, commit, build time |

### Read Operations (sync)

| Command | Description |
|---------|-------------|
| `zaia discover` | List all services in project |
| `zaia discover --service api --include-envs` | Single service detail + env vars |
| `zaia logs --service api [--severity error] [--since 1h] [--limit 100]` | Service logs |
| `zaia validate --file zerops.yml` | Offline YAML validation |
| `zaia validate --content '<yaml>' --type zerops.yml` | Inline YAML validation |
| `zaia search "postgresql connection string" [--limit 5]` | BM25 knowledge search |
| `zaia process <process-id>` | Async process status |
| `zaia env get --service api` | Service env vars |
| `zaia env get --project` | Project env vars |

### Write Operations (async — returns process IDs)

| Command | Description |
|---------|-------------|
| `zaia start --service api` | Start service |
| `zaia stop --service api` | Stop service |
| `zaia restart --service api` | Restart service |
| `zaia scale --service api --min-cpu 1 --max-cpu 5` | Scale service |
| `zaia env set --service api KEY=value` | Set env var |
| `zaia env delete --service api KEY` | Delete env var |
| `zaia import --file services.yml` | Import services from YAML |
| `zaia import --content '<yaml>' --dry-run` | Preview import (sync!) |
| `zaia delete --service api --confirm` | Delete service |
| `zaia subdomain --service api --action enable` | Enable Zerops subdomain |
| `zaia cancel <process-id>` | Cancel process (sync!) |

## Architecture

### Key Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | **No zcli fallback** | Own auth — doesn't read `cli.data` from zcli |
| 2 | **Auto project discovery** | Token must access exactly 1 project (0 → error, 2+ → error) |
| 3 | **Lazy token validation** | Token not validated on login; expiration detected on first API call (401) |
| 4 | **Fire-and-forget async** | Async commands return `processes[]` immediately; MCP layer polls `zaia process` |
| 5 | **JSON-only output** | All stdout is JSON; debug goes to stderr |
| 6 | **Service by hostname** | All commands use `--service <hostname>`, internally resolved to ID |
| 7 | **Idempotent subdomain** | Enable on already-enabled → sync success (not error) |
| 8 | **Import without project:** | `zaia import` only accepts `services:` array; `project:` key → error |

### Response Envelope

Every command returns one of:

```json
{"type":"sync","status":"ok","data":{...}}
{"type":"async","status":"initiated","processes":[{"processId":"...","status":"PENDING"}]}
{"type":"error","code":"SERVICE_NOT_FOUND","error":"...","suggestion":"...","context":{...}}
```

### Error Codes

| Code | Exit | Description |
|------|------|-------------|
| `AUTH_REQUIRED` | 2 | Not authenticated |
| `AUTH_INVALID_TOKEN` | 2 | Invalid token |
| `AUTH_TOKEN_EXPIRED` | 2 | Expired token |
| `TOKEN_NO_PROJECT` | 2 | Token has no project access |
| `TOKEN_MULTI_PROJECT` | 2 | Token has 2+ projects |
| `INVALID_ZEROPS_YML` | 3 | Invalid zerops.yml |
| `INVALID_IMPORT_YML` | 3 | Invalid import.yml |
| `IMPORT_HAS_PROJECT` | 3 | import.yml contains project: section |
| `INVALID_SCALING` | 3 | Invalid scaling parameters |
| `INVALID_PARAMETER` | 3 | Invalid parameter |
| `INVALID_ENV_FORMAT` | 3 | Bad KEY=VALUE format |
| `FILE_NOT_FOUND` | 3 | File doesn't exist |
| `SERVICE_NOT_FOUND` | 4 | Service doesn't exist |
| `PROCESS_NOT_FOUND` | 4 | Process doesn't exist |
| `PROCESS_ALREADY_TERMINAL` | 4 | Process already finished |
| `PERMISSION_DENIED` | 5 | Insufficient permissions |
| `NETWORK_ERROR` | 6 | Network error |
| `INVALID_USAGE` | 3 | Missing command/arg, unknown flag |
| `API_ERROR` | 1 | Generic API error |
| `API_TIMEOUT` | 6 | Timeout |
| `API_RATE_LIMITED` | 6 | Rate limit |

### Code Structure

```
zaia/
├── cmd/zaia/main.go              # Entry point
├── internal/
│   ├── platform/                 # Zerops API abstraction (Client interface, mock, errors)
│   ├── auth/                     # Login/logout, zaia.data storage
│   ├── output/                   # JSON response envelope (Sync/Async/Err)
│   ├── commands/                 # Cobra commands (18 commands)
│   └── knowledge/                # BM25 search engine + 65 embedded docs
├── integration/                  # Multi-command flow tests (StatefulMock)
└── testutil/                     # Golden file + JSON assertion helpers
```

### Knowledge System

- **Engine**: bleve/v2 in-memory BM25 index
- **65 embedded docs** covering all Zerops services, networking, config, operations
- **Field boosts**: title 2.0x, keywords 1.5x, content 1.0x
- **Query expansion**: postgres→postgresql, redis→valkey, mysql→mariadb, etc.
- **URI schema**: `zerops://docs/{category}/{name}`

## Dependencies

```
github.com/blevesearch/bleve/v2     — BM25 full-text search (in-memory)
github.com/spf13/cobra               — CLI framework
github.com/zeropsio/zerops-go v1.0.16 — Zerops API SDK
gopkg.in/yaml.v3                     — YAML parsing
```

## Build & Test

```bash
# Build
go build -o ./zaia ./cmd/zaia

# Run all tests (287+)
go test ./... -count=1

# With race detection
go test ./... -race -count=1

# Integration tests only
go test ./integration/ -v -count=1

# Vet
go vet ./...
```

## Related

- **[ZAIA-MCP](https://github.com/krls2020/zaia-mcp)** — MCP server that calls this CLI as subprocess
- **[Zerops](https://zerops.io)** — Developer-first PaaS with bare-metal infrastructure
