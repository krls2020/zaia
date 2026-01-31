# ZAIA CLI — Development Guide

Go CLI pro AI agenty pracující se Zerops PaaS. Veškerá business logika — auth, discovery, knowledge search, validace, service management. Výstup je vždy JSON.

Součást ZCP: **ZAIA-MCP** (github.com/krls2020/zaia-mcp) volá `zaia` jako subprocess.

> **Public docs:** viz `README.md` (příkazy, error codes, architektura)
> **Design docs:** viz `../design/zaia-cli/` (historický záměr, ne source of truth)

---

## Hierarchie zdrojů pravdy

```
1. Kód (Go types, interface, testy)  ← AUTORITATIVNÍ
2. CLAUDE.md                         ← PROVOZNÍ (workflow, konvence)
3. README.md                         ← PUBLIC DOCS (příkazy, API reference)
4. ../design/zaia-cli/               ← HISTORICKÉ (původní PRD)
```

---

## TDD Workflow

### Povinný workflow pro KAŽDOU změnu

1. **RED**: Napsat failing test PŘED implementací
2. **GREEN**: Minimální implementace
3. **REFACTOR**: Vyčistit, testy zůstávají zelené

### Pravidla

- NIKDY implementace bez odpovídajícího testu
- Table-driven testy (Go idiom)
- Popisné názvy: `TestLogin_SingleProject_Success`
- Max 300 řádků per soubor
- `output.SetWriter` je globální — testy NESMÍ používat `t.Parallel()`
- **Všechny testy mockované** — `platform.MockClient` simuluje Zerops API; E2E testy proti real API žijí v `zaia-mcp/e2e/`

### Příkazy

```bash
go test ./internal/<pkg> -run TestName -v    # Jednotlivý test
go test ./internal/commands -v                # Package
go test ./... -count=1                        # Vše
go test ./... -race -count=1                  # S race detection
go build -o ./zaia ./cmd/zaia                 # Build
go vet ./...                                  # Vet
go test ./integration/ -v -count=1            # Integration
make lint-fast                                # Fast lint (native, ~3s)
make lint-local                               # Full lint (native, ~15s)
make lint                                     # CI lint (3 platformy)
```

---

## Architektura kódu

```
zaia/
├── cmd/zaia/main.go              # Entry point
├── internal/
│   ├── platform/                 # Zerops API: Client interface, Mock, errors, LazyClient
│   ├── auth/                     # Login/logout, zaia.data storage
│   ├── output/                   # JSON response envelope (Sync/Async/Err)
│   ├── commands/                 # Cobra commands (18 commands)
│   └── knowledge/                # BM25 engine + 65 embedded docs
├── integration/                  # Multi-command flow tests (StatefulMock, Harness)
└── testutil/                     # Golden file + JSON assertion helpers
```

### Klíčové soubory (source of truth)

| Soubor | Co definuje |
|--------|-------------|
| `platform/client.go` | Client interface (20+ metod) + všechny typy |
| `platform/errors.go` | Error codes, HTTP mapping, exit codes |
| `output/envelope.go` | Response envelope (sync/async/error) |
| `auth/manager.go` | Login flow, auto project discovery |
| `auth/storage.go` | zaia.data format, OS-specific paths |
| `commands/root.go` | Command wiring, RootDeps, NewRootForTest |
| `knowledge/engine.go` | BM25 store, Search(), List(), Get() |

---

## Konvence

- **JSON-only stdout** — debug na stderr (pokud `--debug`)
- **Service by hostname** — `--service <hostname>`, interně resolví na ID
- **Fire-and-forget async** — async commands vrací `processes[]` okamžitě
- **Bez zcli závislosti** — vlastní auth, nečte `cli.data`
- **Auto project discovery** — token musí mít přístup k právě 1 projektu
- **Import bez project:** — `zaia import` akceptuje pouze `services:` array

---

## Stav implementace

287+ testů, 0 failures. 18 CLI commands. 8 implementačních fází dokončeno.

Zbývá: validace rozšíření (30+ fixture-based testů), hostname regex, debug mode.

---

## Hooks (automatický TDD feedback)

```
.claude/
├── settings.json
└── hooks/
    ├── post-test.sh          # Po Edit/Write .go: go test + go vet + golangci-lint (fast, non-blocking)
    ├── check-claude-md.sh    # Po Edit/Write klíčového souboru: reminder
    └── pre-commit-check.sh   # Před git commit: kontrola CLAUDE.md + golangci-lint (blocking)
```

---

## Údržba

### Kdy aktualizovat CLAUDE.md

| Změna | Akce |
|-------|------|
| Nový command | Aktualizuj "Architektura kódu" strom |
| Nový architektonický vzor | Přidej do "Konvence" |
| Změna stavu implementace | Aktualizuj "Stav implementace" |

Detailní API reference (příkazy, error codes, typy) → viz `README.md` a kód.

### Při aktualizaci knowledge souborů

Knowledge base žije v `../knowledge/` (zdrojová pravda) a je kopírována do `internal/knowledge/embed/` pro Go embed.

```bash
# Automatický sync + testy (doporučeno)
../scripts/sync-knowledge.sh

# Pouze kontrola bez změn (CI/hook)
../scripts/sync-knowledge.sh --check

# Ukázat co by se změnilo
../scripts/sync-knowledge.sh --dry-run
```
