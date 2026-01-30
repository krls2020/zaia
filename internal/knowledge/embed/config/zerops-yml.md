# zerops.yml Specification

## Keywords
zerops.yml, zerops.yaml, build, deploy, run, configuration, pipeline, ports, health check, cron, env variables, prepare commands

## TL;DR
`zerops.yml` (or `zerops.yaml`) defines the build, deploy, and run configuration for each service; the top-level key is the service hostname.

## Structure
```yaml
<service-hostname>:
  build:
    base: <runtime>@<version>
    os: <alpine|ubuntu>                    # optional, default alpine
    prepareCommands:                        # cached in base layer
      - <command>
    buildCommands:                          # main build
      - <command>
    deployFiles:                            # what to deploy
      - <path>
    cache:                                  # what to cache
      - <path>
  run:
    base: <runtime>@<version>              # optional, defaults to build base
    os: <alpine|ubuntu>                    # optional
    prepareCommands:                        # runtime image customization
      - <command>
    addToRunPrepare:                        # persist packages in runtime
      - <command>
    initCommands:                           # run on every container start
      - <command>
    start: <command>                        # single start command
    startCommands:                          # OR multiple start commands
      - <command>
    ports:
      - port: <number>
        protocol: <HTTP|TCP|UDP>
    envVariables:
      KEY: value
    healthCheck:
      httpGet:
        port: <number>
        path: <path>
      # OR
      exec:
        command: <command>
    crontab:
      - command: <command>
        timing: <cron-expression>
    documentRoot: <path>                   # PHP/Nginx/Static only
```

## Key Sections

### build
- `base` — Runtime and version (e.g., `nodejs@22`, `go@1.22`)
- `prepareCommands` — Install system deps; cached in base layer; change invalidates both cache layers
- `buildCommands` — Compile/bundle; runs every build
- `deployFiles` — Files/dirs to deploy (mandatory)
- `cache` — Paths to cache between builds

### run
- `start` / `startCommands` — Entry point (one required)
- `ports` — Internal ports (10-65435, protocols: HTTP/TCP/UDP)
- `healthCheck` — Readiness check (httpGet or exec, 5s timeout, 5min retry window)
- `initCommands` — Runs on every container start (not for package installation)
- `addToRunPrepare` — System packages needed at runtime
- `crontab` — Scheduled tasks (standard cron syntax)
- `envVariables` — Key-value pairs

### Special Features
- `extends` — DRY: inherit from another service config
- `envReplace` — Replace env var references in static files at deploy time
- `routing` — Static/Nginx only: redirects, CORS, custom headers

## Multi-Service Example
```yaml
api:
  build:
    base: nodejs@22
    buildCommands:
      - npm ci && npm run build
    deployFiles: ./dist
  run:
    start: node dist/index.js
    ports:
      - port: 3000
        protocol: HTTP

web:
  build:
    base: nodejs@22
    buildCommands:
      - npm ci && npm run build
    deployFiles: ./build
  run:
    documentRoot: build
```

## Gotchas
1. **Top-level key = service hostname**: Must match the service name in Zerops exactly
2. **`deployFiles` is mandatory**: Build output not auto-deployed — must list explicitly
3. **`initCommands` runs every restart**: Don't install packages here — use `prepareCommands`
4. **`prepareCommands` change = full rebuild**: Both cache layers invalidated
5. **Ports 80/443 reserved**: Cannot use them — Zerops handles SSL termination
6. **`start` vs `startCommands`**: Use `start` for single process, `startCommands` for multiple

## See Also
- zerops://config/import-yml
- zerops://platform/build-cache
- zerops://services/_common-runtime
- zerops://examples/zerops-yml-runtimes
