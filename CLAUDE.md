# ZAIA CLI — Isolated Development Guide

> Toto je **self-contained** reference pro vývoj ZAIA CLI.
> Claude Code se spouští přímo v tomto adresáři (`./zaia/`).

---

## ⚠️ POVINNÉ: Údržba CLAUDE.md

**CLAUDE.md MUSÍ být vždy aktuální.** Toto je jediný zdroj pravdy pro izolovaný vývoj ZAIA.

### Kdy aktualizovat

| Změna | Co aktualizovat v CLAUDE.md |
|-------|----------------------------|
| Nový/změněný interface v `platform/client.go` | Sekce **Klíčové typy a interface** |
| Nový command v `commands/` | Sekce **Příkazy ZAIA CLI** + **Architektura kódu** |
| Nový error code v `platform/errors.go` | Sekce **Error codes a exit codes** |
| Změna response formátu v `output/envelope.go` | Sekce **Response Envelope** |
| Změna auth flow v `auth/manager.go` | Sekce **Architektonická rozhodnutí** |
| Nová závislost v `go.mod` | Sekce **Závislosti** |
| Dokončení implementační fáze | Sekce **Aktuální stav implementace** |
| Nový architektonický vzor | Sekce **Klíčová architektonická rozhodnutí** |
| Změna knowledge engine | Sekce **Knowledge system** |
| Nový test pattern | Sekce **Vzory pro psaní testů** |

### Jak aktualizovat

1. **Po každé změně klíčového souboru** → zkontroluj relevantní sekci CLAUDE.md
2. **Po dokončení úkolu/fáze** → aktualizuj sekci "Aktuální stav implementace"
3. **Při přidání nového konceptu** → přidej novou sekci nebo rozšiř existující
4. **Při změně cest/struktury** → aktualizuj všechny reference

**Neaktuální CLAUDE.md = agent pracuje se špatnými informacemi.**

### Automatické vynucení (hooks)

Hooks v `.claude/settings.json` automaticky:
- **Po Edit/Write Go souboru** → spustí `go test` pro změněný package (TDD feedback)
- **Po Edit/Write klíčového souboru** → zobrazí připomínku aktualizace CLAUDE.md
- **Před `git commit`** → zkontroluje zda CLAUDE.md je ve staged area pokud se změnily klíčové soubory

---

## Co je ZAIA

**ZAIA** je Go CLI binárka pro AI agenty pracující se Zerops PaaS. Obsahuje veškerou business logiku — autentizaci, správu služeb, BM25 knowledge search, validaci YAML. Výstup je vždy JSON.

ZAIA je součástí třísložkového systému:

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│  ZAIA-MCP       │────▶│  ZAIA CLI    │     │  ZCLI       │
│  (MCP server)   │     │  (tento repo)│     │  (produkční)│
│  tenká vrstva   │────▶│  business    │     │  deploy,push│
└─────────────────┘     │  logika      │     └─────────────┘
                        └──────────────┘
                              │
                              ▼
                     ┌─────────────────┐
                     │ Zerops Cloud API│
                     └─────────────────┘
```

- **ZAIA CLI** (tento projekt): Veškerá logika, knowledge, validace, auth
- **ZAIA-MCP** (dosud neimplementováno): Tenká MCP vrstva volající `zaia` jako subprocess
- **ZCLI**: Produkční CLI — ZAIA ho nevolá, nezávisí na něm

---

## Nadřazený repozitář

Tento projekt žije v `./zaia/` podadresáři repo `/Users/macbook/Documents/Zerops-MCP/`. Nadřazený adresář obsahuje:

| Cesta (relativní k `../`) | Obsah |
|---|---|
| `spec/zaia-cli/` | **Autoritativní specifikace** — 8 souborů, ~7,500 řádků |
| `spec/zaia-mcp/` | ZAIA-MCP system prompt spec |
| `docs/decisions/` | Architektonická rozhodnutí (5 souborů) |
| `docs/context/` | Zerops kontext, vize, analýzy |
| `knowledge/` | **Zdrojové knowledge soubory** (65 MD, ~4,800 řádků) |
| `CLAUDE.md` | Nadřazené projektové instrukce |

### Specifikační dokumenty (`../spec/zaia-cli/`)

| Soubor | Co definuje |
|--------|-------------|
| `README.md` | Přehled architektury, komponenty |
| `auth.md` | **Login flow, auto project discovery, lazy token validation, storage format** |
| `project-scope.md` | **Single-project architektura, token musí mít přístup k 1 projektu** |
| `commands.md` | **Všechny CLI příkazy s flagy, response typy, edge cases** |
| `output.md` | **Response envelope (sync/async/error), process status mapping, error codes** |
| `validation.md` | **YAML validace zerops.yml + import.yml, pravidla, příklady** |
| `tools.md` | **Platform Client interface (20+ metod), všechny datové typy** |
| `knowledge.md` | **BM25 search, URI schema, query expansion, MCP Resources** |

### Design decisions (`../docs/decisions/`)

| Soubor | Rozhodnutí |
|--------|-----------|
| `architecture.md` | ZAIA vs ZAIA-MCP vs ZCLI — proč tři komponenty |
| `design-principles.md` | 7 principů: happy path first, opinionated defaults, token efficiency |
| `knowledge-system.md` | Proč BM25 (ne embeddings), embedded v binary |
| `subagent-system.md` | Orchestrator-Worker model, transcript debugging |
| `system-prompt.md` | ZAIA-MCP prompt design (~250 tokenů) |

**Při implementaci VŽDY čti příslušný spec soubor.** Spec je autoritativní — kód musí odpovídat spec.

---

## Vývoj: TDD Workflow

### Povinný workflow pro KAŽDOU změnu

1. **RED**: Napsat failing test PŘED implementací
   - `go test ./internal/<pkg> -run TestNew` → FAIL
   - Test MUSÍ failovat ze správného důvodu (ne compile error)
2. **GREEN**: Napsat minimální implementaci
   - Jen to co je potřeba aby test prošel
   - Žádné extra features, žádný over-engineering
3. **REFACTOR**: Vyčistit kód, testy zůstávají zelené
   - `go test ./internal/... -count=1` → PASS

### Pravidla

- NIKDY nepsat implementaci bez odpovídajícího testu
- Při modifikaci existující funkce: ověřit že test existuje, pokud ne → napsat
- Table-driven testy preferovat (Go idiom)
- Jeden test = jedno chování
- Popisné názvy testů: `TestLogin_SingleProject_Success`
- Max 300 řádků per file — LLM drží celý soubor v kontextu
- Po každé změně: `go test ./internal/<pkg> -v`

### Příkazy

```bash
# Jednotlivý test
go test ./internal/auth -run TestLogin_Single -v

# Package
go test ./internal/commands -v

# Vše
go test ./... -count=1

# S race detection
go test ./... -race -count=1

# Build
go build -o ./zaia ./cmd/zaia

# Vet
go vet ./...

# Integration testy
go test ./integration/ -v -count=1

# Konkrétní flow
go test ./integration/ -run TestFlow_EnvSetThenGet -v
```

### Hooks (automatický TDD feedback)

Hooks jsou v `.claude/settings.json` a `.claude/hooks/`:

```
.claude/
├── settings.json               # Hook konfigurace
└── hooks/
    ├── post-test.sh            # Po Edit/Write: go test + go vet pro změněný package
    ├── check-claude-md.sh      # Po Edit/Write: připomínka aktualizace CLAUDE.md
    └── pre-commit-check.sh     # Před git commit: kontrola CLAUDE.md ve staged area
```

| Hook | Trigger | Co dělá |
|------|---------|---------|
| `post-test.sh` | PostToolUse (Edit\|Write) na `.go` soubor | Spustí `go test` + `go vet` pro package změněného souboru |
| `check-claude-md.sh` | PostToolUse (Edit\|Write) na klíčový soubor | Zobrazí připomínku aktualizace CLAUDE.md |
| `pre-commit-check.sh` | PreToolUse (Bash) na `git commit` | Zkontroluje zda CLAUDE.md je staged pokud se změnily klíčové soubory |

**Klíčové soubory** (spouští check-claude-md):
`platform/client.go`, `commands/root.go`, `go.mod`, `knowledge/engine.go`, `output/envelope.go`, `auth/manager.go`, `auth/storage.go`, `platform/errors.go`

Hooks čtou JSON ze stdin (Claude Code hook protocol):
```json
{
  "tool_name": "Edit",
  "tool_input": {"file_path": "/abs/path/to/file.go", ...},
  "tool_response": {...}
}
```

---

## Architektura kódu

```
zaia/
├── cmd/zaia/main.go                    # Entry point, commands.Execute() wrapper, exit codes
├── go.mod                              # github.com/zeropsio/zaia
│
├── internal/
│   ├── platform/                       # Zerops API abstrakce
│   │   ├── client.go                   # Client interface (20+ metod) + všechny typy
│   │   ├── zerops.go                   # ⭐ Reálná implementace Client (zerops-go SDK)
│   │   ├── zerops_test.go              # Testy error mapping, constructor
│   │   ├── logfetcher.go              # LogFetcher — HTTP log backend klient
│   │   ├── logfetcher_test.go         # httptest testy pro log fetcher
│   │   ├── lazy.go                    # LazyClient — inicializace při prvním API callu
│   │   ├── lazy_test.go               # Testy lazy init, concurrency
│   │   ├── mock.go                     # Fluent mock pro testy
│   │   ├── mock_test.go
│   │   ├── errors.go                   # Error codes, HTTP mapping, exit codes, PlatformError
│   │   └── errors_test.go
│   │
│   ├── auth/                           # Autentizace
│   │   ├── manager.go                  # Login (auto project discovery), Logout, GetCredentials
│   │   ├── manager_test.go
│   │   ├── storage.go                  # zaia.data — atomic file I/O, OS-specific paths
│   │   └── storage_test.go
│   │
│   ├── output/                         # JSON response envelope
│   │   ├── envelope.go                 # Sync(), Async(), Err() builders
│   │   ├── envelope_test.go
│   │   └── errors.go                   # ZaiaError (implements error + ExitCode)
│   │
│   ├── commands/                       # Cobra commands (thin wrappers)
│   │   ├── root.go                     # NewRoot() — wiring, version ldflags
│   │   ├── helpers.go                  # resolveCredentials, findServiceByHostname
│   │   ├── login.go                    # zaia login <token> [--url]
│   │   ├── logout.go                   # zaia logout
│   │   ├── status.go                   # zaia status (no API call)
│   │   ├── version.go                  # zaia version
│   │   ├── discover.go                 # zaia discover [--service] [--include-envs]
│   │   ├── process.go                  # zaia process <id>
│   │   ├── cancel.go                   # zaia cancel <id>
│   │   ├── logs.go                     # zaia logs --service <name> [--severity] [--since]
│   │   ├── validate.go                 # zaia validate --file|--content [--type]
│   │   ├── search.go                   # zaia search <query> [--limit]
│   │   ├── manage.go                   # zaia start|stop|restart|scale --service <name>
│   │   ├── env.go                      # zaia env get|set|delete --service|--project
│   │   ├── importcmd.go               # zaia import --file|--content [--dry-run]
│   │   ├── delete.go                   # zaia delete --service <name> --confirm
│   │   ├── subdomain.go               # zaia subdomain --service --action enable|disable
│   │   ├── testhelpers_test.go         # Shared test setup
│   │   └── *_test.go                   # Testy pro každý command
│   │
│   └── knowledge/                      # BM25 knowledge engine
│       ├── engine.go                   # Store, Search(), List(), Get(), Provider interface
│       ├── engine_test.go              # 32 testů — search queries, hit rate, parsing
│       ├── documents.go                # go:embed loader, parseDocument(), URI conversion
│       ├── query.go                    # expandQuery(), suggestions, snippet extraction
│       └── embed/                      # 65 embedded knowledge MD files
│           ├── config/                 # zerops-yml, import-yml, deploy-patterns...
│           ├── decisions/              # choose-database, choose-cache...
│           ├── services/               # nodejs, postgresql, valkey, nginx...
│           ├── operations/             # logging, ci-cd, production-checklist...
│           ├── platform/               # scaling, backup, env-variables...
│           ├── networking/             # public-access, cloudflare, firewall...
│           ├── examples/               # connection-strings, zerops-yml-runtimes
│           └── gotchas/                # common (37 gotchas)
│
├── integration/                        # ⭐ Integration testy (multi-command flows)
│   ├── harness.go                     # Harness struct — Run("discover --service api") → Result
│   ├── stateful_mock.go               # StatefulMock — wraps platform.Client, trackuje mutace
│   ├── fixtures.go                    # Reusable project states (empty, full, stopped, unauth)
│   ├── cli_wiring_test.go            # Všech 18 commands registrováno v root
│   ├── flow_auth_test.go             # login → discover → logout → discover fails
│   ├── flow_lifecycle_test.go        # discover → start → process → discover (status changed)
│   ├── flow_env_test.go              # env set → env get → env delete → env get
│   ├── flow_import_delete_test.go    # import → discover → delete
│   ├── flow_error_propagation_test.go # AUTH_REQUIRED, SERVICE_NOT_FOUND pro všechny commands
│   ├── envelope_test.go              # sync/async/error JSON format + exit codes
│   └── flow_subdomain_test.go        # subdomain enable → idempotent enable → disable
│
└── testutil/
    └── golden.go                       # AssertGolden(), AssertJSONEqual()
```

---

## Klíčové typy a interface

### RootDeps + NewRootForTest (`internal/commands/root.go`)

```go
type RootDeps struct {
    StoragePath   string
    Client        platform.Client
    ClientFactory func(token, apiHost string) platform.Client
    LogFetcher    platform.LogFetcher
}

func NewRootForTest(deps RootDeps) *cobra.Command
```

Used by integration tests to inject mocked dependencies. Mirrors `NewRoot()` but with all deps injectable.

### Platform Client (`internal/platform/client.go`)

```go
type Client interface {
    GetUserInfo(ctx context.Context) (*UserInfo, error)
    ListProjects(ctx context.Context, clientID string) ([]Project, error)
    GetProject(ctx context.Context, projectID string) (*Project, error)
    ListServices(ctx context.Context, projectID string) ([]ServiceStack, error)
    GetService(ctx context.Context, serviceID string) (*ServiceStack, error)
    StartService(ctx context.Context, serviceID string) (*Process, error)
    StopService(ctx context.Context, serviceID string) (*Process, error)
    RestartService(ctx context.Context, serviceID string) (*Process, error)
    SetAutoscaling(ctx context.Context, serviceID string, params AutoscalingParams) (*Process, error)
    GetServiceEnv(ctx context.Context, serviceID string) ([]EnvVar, error)
    SetServiceEnvFile(ctx context.Context, serviceID string, content string) (*Process, error)
    DeleteUserData(ctx context.Context, serviceID, envID string) (*Process, error)
    GetProjectEnv(ctx context.Context, projectID string) ([]EnvVar, error)
    CreateProjectEnv(ctx context.Context, projectID, key, value string) (*Process, error)
    DeleteProjectEnv(ctx context.Context, projectID, envID string) (*Process, error)
    ImportServices(ctx context.Context, projectID, yamlContent string) (*ImportResult, error)
    DeleteService(ctx context.Context, serviceID string) (*Process, error)
    GetProcess(ctx context.Context, processID string) (*Process, error)
    CancelProcess(ctx context.Context, processID string) (*Process, error)
    EnableSubdomainAccess(ctx context.Context, serviceID string) (*Process, error)
    DisableSubdomainAccess(ctx context.Context, serviceID string) (*Process, error)
    GetProjectLog(ctx context.Context, projectID string) (*LogAccess, error)
}

type LogFetcher interface {
    FetchLogs(ctx context.Context, access *LogAccess, params LogFetchParams) ([]LogEntry, error)
}
```

### Response Envelope (`internal/output/envelope.go`)

Každý CLI příkaz vrací jedno z:

```json
// Sync (okamžitá data)
{"type":"sync","status":"ok","data":{...}}

// Async (vrací process IDs, fire-and-forget)
{"type":"async","status":"initiated","processes":[{"processId":"...","status":"PENDING",...}]}

// Error
{"type":"error","code":"SERVICE_NOT_FOUND","error":"...","suggestion":"...","context":{...}}
```

### Mock (`internal/platform/mock.go`)

Fluent API pro testy:

```go
mock := platform.NewMock().
    WithProjects([]platform.Project{{ID: "p1", Name: "app"}}).
    WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}}).
    WithError("StartService", fmt.Errorf("service stopped"))
```

### Knowledge Store (`internal/knowledge/engine.go`)

```go
store := knowledge.GetEmbeddedStore()           // singleton
results := store.Search("postgresql port", 5)   // BM25 search
doc, _ := store.Get("zerops://docs/services/postgresql") // by URI
resources := store.List()                        // all 65 docs
suggestions := store.GenerateSuggestions(query, results)
```

---

## Aktuální stav implementace

### Hotovo (Fáze 0–8)

| Fáze | Co | Testy |
|------|----|-------|
| 0 | Scaffold (go.mod, main.go, root.go) | — |
| 1 | Platform Client interface, typy, Mock, Output envelope | 48 |
| 2 | Auth (login/logout/status/version) | 17 |
| 3 | Discovery + Process + Cancel | 11 |
| 4–5 | Logs, Validate, Manage, Env, Import, Delete, Subdomain | 17 |
| 6 | Knowledge BM25 search (65 docs, bleve index) | 32 |
| 7 | Real API Client (zerops.go, logfetcher.go, lazy.go, root wiring) | 121 |
| 8 | Integration testy (multi-command flows, stateful mock, envelope validation) | 41 funkcí (~69 test cases) |

**Celkem: 287+ testů, 0 failures.**

### Fáze 7: Real API Client (dokončeno)

| Soubor | Popis |
|--------|-------|
| `platform/zerops.go` | ⭐ Reálná implementace Client interface (21 metod) pomocí zerops-go SDK v1.0.16 |
| `platform/logfetcher.go` | HTTP klient pro Zerops log backend (step 2 of 2-step log retrieval) |
| `platform/lazy.go` | LazyClient — proxy Client, inicializuje ZeropsClient při prvním API callu |
| `commands/root.go` | Napojení: LazyClient + NewLogFetcher() + `NewRootForTest()` pro integrační testy |
| `platform/errors.go` | Přidán PlatformError typ pro SDK error mapping |

### Zbývá implementovat

| Úkol | Popis | Priorita |
|------|-------|----------|
| **~~E2E testy~~** | ✅ Implementováno jako `integration/` package (StatefulMock, multi-command flows) | Hotovo |
| **Validace rozšíření** | 30+ fixture-based testů pro zerops.yml/import.yml (viz `../spec/zaia-cli/validation.md`) | Střední |
| **Hostname validace** | Regex `^[a-z0-9][a-z0-9-]*[a-z0-9]$` v commands | Nízká |
| **Debug mode** | `--debug` flag → stderr logging | Nízká |
| **ZAIA-MCP server** | Separátní binárka volající `zaia` CLI jako subprocess | Samostatný projekt |

---

## Klíčová architektonická rozhodnutí

### 1. Bez zcli fallbacku

ZAIA má vlastní autentizaci. Nečte `cli.data` z zcli. Token výhradně přes `zaia login`.

### 2. Auto project discovery

Při `zaia login <token>` se volá `ListProjects()`. Token musí mít přístup k právě 1 projektu:
- 0 projektů → `TOKEN_NO_PROJECT`
- 2+ projektů → `TOKEN_MULTI_PROJECT`
- 1 projekt → uložit project ID + name do `zaia.data`

### 3. Lazy token validation

Token se nevaliduje při login (jen se uloží + discover project). Expirace se detekuje až při API callu (HTTP 401 → `AUTH_TOKEN_EXPIRED`).

### 4. Fire-and-forget async

Async commands (start/stop/restart/scale/import/delete/subdomain/env set/env delete) vrací `{"type":"async","processes":[...]}` okamžitě. Žádný `--wait` flag. MCP vrstva polluje `zaia process <id>`.

### 5. JSON-only output

Veškerý stdout je JSON. Debug info jde na stderr (pokud `--debug`).

### 6. Service resolution by hostname

Všechny commands používají `--service <hostname>` (ne ID). Interně se hostname resolví na ID přes `ListServices()`.

### 7. Idempotentní subdomain

`zaia subdomain --action enable` na již enabled službu → sync response `{"status":"already_enabled"}` (ne error).

### 8. JSON error output pro ALL error paths

Cobra errors (unknown command, missing args, unknown flags) are caught by `commands.Execute()` wrapper and converted to JSON error envelopes. Parent commands without subcommands (root, env) have `RunE` that returns JSON. `MarkFlagRequired` is NOT used — all flag validation happens in `RunE` to ensure JSON output.

### 9. Import bez project: sekce

`zaia import` akceptuje YAML pouze se `services:` array. Pokud obsahuje `project:` → error `IMPORT_HAS_PROJECT`.

---

## Error codes a exit codes

| Code | Exit | Popis |
|------|------|-------|
| `AUTH_REQUIRED` | 2 | Není přihlášen |
| `AUTH_INVALID_TOKEN` | 2 | Neplatný token |
| `AUTH_TOKEN_EXPIRED` | 2 | Expirovaný token |
| `TOKEN_NO_PROJECT` | 2 | Token nemá přístup k žádnému projektu |
| `TOKEN_MULTI_PROJECT` | 2 | Token má přístup k 2+ projektům |
| `INVALID_ZEROPS_YML` | 3 | Nevalidní zerops.yml |
| `INVALID_IMPORT_YML` | 3 | Nevalidní import.yml |
| `IMPORT_HAS_PROJECT` | 3 | import.yml obsahuje project: sekci |
| `INVALID_SCALING` | 3 | Nevalidní scaling parametry |
| `INVALID_PARAMETER` | 3 | Nevalidní parametr |
| `INVALID_ENV_FORMAT` | 3 | Špatný KEY=VALUE formát |
| `FILE_NOT_FOUND` | 3 | Soubor neexistuje |
| `SERVICE_NOT_FOUND` | 4 | Služba neexistuje |
| `PROCESS_NOT_FOUND` | 4 | Proces neexistuje |
| `PROCESS_ALREADY_TERMINAL` | 4 | Proces již dokončen (nelze cancel) |
| `PERMISSION_DENIED` | 5 | Nedostatečná oprávnění |
| `NETWORK_ERROR` | 6 | Síťová chyba |
| `INVALID_USAGE` | 3 | Chybějící příkaz/argument, neznámý command/flag |
| `API_ERROR` | 1 | Obecná API chyba |
| `API_TIMEOUT` | 6 | Timeout |
| `API_RATE_LIMITED` | 6 | Rate limit |

---

## Závislosti

```
github.com/blevesearch/bleve/v2    — BM25 full-text search (in-memory)
github.com/spf13/cobra              — CLI framework
github.com/zeropsio/zerops-go v1.0.16 — Zerops API SDK (reálný klient)
gopkg.in/yaml.v3                    — YAML parsing
```

---

## Příkazy ZAIA CLI

### Auth & Info

| Příkaz | Response | Popis |
|--------|----------|-------|
| `zaia login <token>` | sync | Auto project discovery, uložení do zaia.data |
| `zaia login <token> --url api.staging.zerops.io` | sync | Login na staging |
| `zaia logout` | sync | Smaže zaia.data |
| `zaia status` | sync | Zobrazí přihlášeného uživatele + projekt (bez API callu) |
| `zaia version` | sync | Verze, commit, build time |

### Read Operations (sync)

| Příkaz | Popis |
|--------|-------|
| `zaia discover` | List všech služeb v projektu |
| `zaia discover --service api --include-envs` | Detail jedné služby + env vars |
| `zaia logs --service api [--severity error] [--since 1h] [--limit 100]` | Logy služby |
| `zaia validate --file zerops.yml` | Offline YAML validace |
| `zaia validate --content '<yaml>' --type zerops.yml` | Inline YAML validace |
| `zaia search "postgresql connection string" [--limit 5]` | BM25 knowledge search |
| `zaia process <process-id>` | Stav async procesu |
| `zaia env get --service api` | Env vars služby |
| `zaia env get --project` | Env vars projektu |

### Write Operations (async — vrací process IDs)

| Příkaz | Popis |
|--------|-------|
| `zaia start --service api` | Start služby |
| `zaia stop --service api` | Stop služby |
| `zaia restart --service api` | Restart služby |
| `zaia scale --service api --min-cpu 1 --max-cpu 5` | Scaling |
| `zaia env set --service api KEY=value` | Set env var |
| `zaia env delete --service api KEY` | Delete env var |
| `zaia import --file services.yml` | Import služeb z YAML |
| `zaia import --content '<yaml>' --dry-run` | Preview importu (sync!) |
| `zaia delete --service api --confirm` | Smazání služby |
| `zaia subdomain --service api --action enable` | Enable Zerops subdomény |
| `zaia cancel <process-id>` | Zrušení procesu (sync!) |

---

## Knowledge system

### Embedded docs

65 markdown souborů v `internal/knowledge/embed/` — kopie z `../knowledge/`. Při aktualizaci knowledge souborů je třeba překopírovat:

```bash
cp -R ../knowledge/* internal/knowledge/embed/
```

### BM25 Search

- **Engine**: bleve/v2 in-memory index
- **Field boosts**: title 2.0x, keywords 1.5x, content 1.0x
- **Query expansion**: postgres→postgresql, redis→valkey, mysql→mariadb, node→nodejs, ssl→tls, env→environment variable, ...
- **Hit rate**: Hit@1 62%, Hit@3 100%
- **URI schema**: `zerops://docs/{category}/{name}` (bez .md)

### Document format

Každý knowledge soubor:

```markdown
# {Topic} on Zerops

## Keywords
keyword1, keyword2, alias1, alias2

## TL;DR
Jedna věta — nejdůležitější Zerops-specific fakt.

## Zerops-Specific Behavior
- Jen co LLM neví
- Porty, env vars, cesty, defaulty, limity

## Configuration
Copy-paste ready YAML/config

## Gotchas
1. **Problem**: Solution

## See Also
- zerops://docs/path/to/related
```

### Unsupported services

Při hledání mongodb, dynamodb, kubernetes, memcached, sqlite → vrací suggestions s alternativami.

---

## Vzory pro psaní testů

### Table-driven (primární)

```go
func TestMyFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantErr  bool
        wantCode string
    }{
        {"valid", "good input", false, ""},
        {"invalid", "bad input", true, "INVALID_PARAMETER"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test body
        })
    }
}
```

### Command test s mock

```go
func TestMyCmd_Success(t *testing.T) {
    storagePath := setupAuthenticatedStorage(t)
    mock := platform.NewMock().
        WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

    cmd := NewMyCommand(storagePath, mock)
    cmd.SetArgs([]string{"--service", "api"})

    var stdout bytes.Buffer
    output.SetWriter(&stdout)
    defer output.ResetWriter()

    err := cmd.Execute()
    if err != nil {
        t.Fatal(err)
    }

    var resp map[string]interface{}
    json.Unmarshal(stdout.Bytes(), &resp)
    // assertions...
}
```

### setupAuthenticatedStorage helper

Definován v `discover_test.go` — vytváří temp dir s validním `zaia.data`:

```go
func setupAuthenticatedStorage(t *testing.T) string {
    t.Helper()
    dir := t.TempDir()
    storage := auth.NewStorage(dir)
    storage.Save(auth.Data{
        Token:   "test-token",
        APIHost: "https://api.zerops.io",
        Project: auth.ProjectInfo{ID: "proj-1", Name: "test-project"},
    })
    return dir
}
```

### Output capture

```go
var stdout bytes.Buffer
output.SetWriter(&stdout)
defer output.ResetWriter()

// ... execute command ...

var resp map[string]interface{}
json.Unmarshal(stdout.Bytes(), &resp)
```

### Integration test s Harness (multi-command flows)

```go
func TestFlow_MyFeature(t *testing.T) {
    h := integration.NewHarness(t)
    integration.FixtureFullProject(h)       // authenticated, 3 services

    r := h.MustRun("start --service api")   // execute CLI command
    r.AssertType("async")                   // verify response type
    r.AssertExitCode(0)

    r = h.MustRun("discover")              // verify side effects
    data := r.Data()
    services := data["services"].([]interface{})
    svc := services[0].(map[string]interface{})
    if svc["status"] != "ACTIVE" {
        t.Errorf("expected ACTIVE, got %v", svc["status"])
    }
}
```

**Harness** vytváří čerstvý `NewRootForTest()` pro každé `Run()` (simuluje nové CLI volání).
**StatefulMock** trackuje mutace — write operace ovlivní následné read operace.
**Fixtures**: `FixtureUnauthenticated`, `FixtureEmptyProject`, `FixtureFullProject`, `FixtureStoppedService`.

**DŮLEŽITÉ**: `output.SetWriter` je globální — testy NESMÍ používat `t.Parallel()`.

---

## Process status mapping

API status → ZAIA status:

| API | ZAIA | Popis |
|-----|------|-------|
| `PENDING` | `PENDING` | Čeká |
| `RUNNING` | `RUNNING` | Běží |
| `DONE` | `FINISHED` | Úspěšně dokončen |
| `FAILED` | `FAILED` | Selhal |
| `CANCELLED` | `CANCELED` | Zrušen (1 L) |

---

## Scaling parametry

`zaia scale` podporuje:

| Flag | Typ | Popis |
|------|-----|-------|
| `--service` | string | Hostname služby (required) |
| `--cpu-mode` | string | SHARED nebo DEDICATED |
| `--min-cpu` | int | Min CPU cores |
| `--max-cpu` | int | Max CPU cores |
| `--min-ram` | float | Min RAM v GB |
| `--max-ram` | float | Max RAM v GB |
| `--min-disk` | float | Min disk v GB |
| `--max-disk` | float | Max disk v GB |
| `--start-containers` | int | Počáteční repliky |
| `--min-containers` | int | Min repliky (horizontální scaling) |
| `--max-containers` | int | Max repliky |

Validace: min <= max pro CPU, RAM, disk, containers. Alespoň jeden parametr musí být zadán.

---

## Env var format

`zaia env set --service api KEY=value ANOTHER=value2`

- Split na prvním `=` (hodnota může obsahovat `=`)
- Prázdná hodnota je validní: `KEY=`
- Bez `=` → error `INVALID_ENV_FORMAT`
- `zaia env set --project KEY=value` → project scope

---

## Storage format (`zaia.data`)

JSON soubor s OS-specific umístěním:

| OS | Cesta |
|----|-------|
| macOS | `~/Library/Application Support/zerops/zaia.data` |
| Linux | `~/.config/zerops/zaia.data` |
| Override | `ZAIA_DATA_FILE_PATH` env var |

```json
{
  "token": "...",
  "apiHost": "https://api.zerops.io",
  "regionData": {"name": "default", "address": "https://api.zerops.io", "isDefault": true},
  "project": {"id": "proj-uuid", "name": "my-app"},
  "user": {"name": "John", "email": "john@example.com"}
}
```

Atomic write: zapis do `.new`, pak `os.Rename()`. Permissions: `0600`.

---

## Referenční implementace: mcp60

Projekt `/Users/macbook/Sites/mcp60` obsahuje historickou MCP server implementaci. Klíčové patterny:

| mcp60 soubor | Co obsahuje | ZAIA ekvivalent |
|---|---|---|
| `src/tools/discovery.go` | `sdk.GetAllServices()` | `commands/discover.go` |
| `src/tools/manageLifecycle.go` | Start/stop/restart | `commands/manage.go` |
| `src/tools/scaleService.go` | Scaling | `commands/manage.go` (scale) |
| `src/tools/setEnv.go` | KEY=value parsing | `commands/env.go` |
| `src/tools/import.go` | YAML import | `commands/importcmd.go` |
| `src/tools/getServiceLogs.go` | Log retrieval | `commands/logs.go` |
| `src/tools/controlSubdomain.go` | Enable/disable | `commands/subdomain.go` |
| `src/zeropsSdk/handler.go` | zerops-go SDK wrapper | `platform/zerops.go` (TODO) |
| `src/knowledge/store.go` | `embed.FS` knowledge loading | `knowledge/documents.go` |

### Zerops-go SDK patterny (pro implementaci `platform/zerops.go`)

```go
// Typed IDs
uuid.ServiceStackId(serviceID)
uuid.ProjectId(projectID)
uuid.ProcessId(processID)
path.ProjectId{Id: uuid.ProjectId(id)}

// Status checking
process.Status.Is(enum.ProcessStatusEnumFinished)
process.Status.Is(enum.ProcessStatusEnumFailed)

// ID extraction
service.Id.TypedString().String()
```

---

## Zerops kontext (zkrácený)

**Zerops** je developer-first PaaS:
- Full Linux kontejnery (Incus-based), plný SSH přístup
- Managed services: PostgreSQL, MariaDB, Valkey (Redis-kompatibilní), Elasticsearch, Meilisearch, Kafka, NATS, S3
- VXLAN networking — izolovaná síť per projekt
- Auto-scaling (horizontální + vertikální)
- Deploy přes `zcli push` (ne přes ZAIA)

### Klíčové Zerops specifika

- **Valkey místo Redis**: Redis = Valkey na Zeropsu
- **MariaDB místo MySQL**: MySQL-kompatibilní
- **Bez MongoDB**: Neexistuje na Zeropsu
- **Bez Kubernetes**: PaaS, ne container orchestrator
- **Import YAML = services only**: `project:` sekce je jen pro plný import (mimo ZAIA scope)
- **HA mode je immutable**: Nelze změnit po vytvoření služby
- **Porty**: PostgreSQL 5432 (primary), 5433 (replicas), 6432 (external TLS)

---

## Údržba

### Při aktualizaci knowledge souborů

```bash
# Z nadřazeného adresáře:
cp -R ../knowledge/* internal/knowledge/embed/
go test ./internal/knowledge/ -v
```

### Při přidání nového příkazu

1. Vytvořit `commands/newcmd.go` + `commands/newcmd_test.go`
2. Přidat do `commands/root.go` → `rootCmd.AddCommand(NewXxx(...))`
3. Zkontrolovat spec: `../spec/zaia-cli/commands.md`
4. `go test ./internal/commands/ -v`

### Při změně Platform Client interface

1. Upravit `platform/client.go`
2. Aktualizovat `platform/mock.go` (musí implementovat interface)
3. Compile-time check: `var _ Client = (*Mock)(nil)`
4. Aktualizovat relevantní command a test
5. `go test ./... -count=1`

### Při změně error codes

1. Přidat konstantu v `platform/errors.go`
2. Přidat exit code mapping v `ExitCodeForError()`
3. Přidat test case v `errors_test.go`
4. `go test ./internal/platform/ -v`
