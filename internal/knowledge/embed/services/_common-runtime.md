# Common Runtime Patterns on Zerops

## Keywords
runtime, build, deploy, run, working directory, port, prepare commands, build commands, deploy files, common runtime, pipeline

## TL;DR
All Zerops runtimes share the same build pipeline (prepare → build → deploy), use `/var/www` as working directory, and must define ports in the 10-65435 range.

## Build Pipeline
1. **prepareCommands** — Install OS dependencies, cached in base layer
2. **buildCommands** — Compile/bundle application
3. **deployFiles** — Files to deploy to runtime container

## Working Directory
- Build: `/var/www` (source code mounted here)
- Runtime: `/var/www` (deployed files placed here)

## Port Configuration
- Range: 10-65435 (80 and 443 reserved by Zerops)
- Define in `zerops.yaml` under `run.ports`
- Protocols: HTTP, TCP, UDP

## Shared Patterns

### Build Cache
```yaml
build:
  cache:
    - node_modules      # Node.js
    - .next/cache       # Next.js
    - vendor            # PHP Composer
    - __pycache__       # Python
    - target            # Rust/Java
```

### Health Check
```yaml
run:
  healthCheck:
    httpGet:
      port: 3000
      path: /health
```

### Environment Variables
```yaml
run:
  envVariables:
    NODE_ENV: production
    PORT: "3000"
```

## Gotchas
1. **initCommands run every restart**: Use `prepareCommands` for package installation — `initCommands` runs on every container start
2. **deployFiles is mandatory**: Build output not automatically deployed — must explicitly list files/dirs
3. **Port 80/443 reserved**: Your app must use another port — Zerops handles SSL on 80/443

## See Also
- zerops://config/zerops-yml
- zerops://platform/build-cache
- zerops://platform/scaling
- zerops://platform/infrastructure
