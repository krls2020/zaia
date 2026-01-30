# Deno on Zerops

## Keywords
deno, javascript, typescript, deno deploy, permissions, deno.json, deno task, secure runtime

## TL;DR
Deno on Zerops requires explicit permission flags (`--allow-net`, `--allow-env`, etc.); uses `deno.json` for tasks, and has native TypeScript support without a build step.

## Zerops-Specific Behavior
- Versions: 1+
- Base: Alpine (default)
- Config: `deno.json` (tasks, imports)
- Working directory: `/var/www`
- No default port — must configure
- npm compatibility: `npm:` specifier for npm packages
- Permission model: Explicit flags required

## Configuration
```yaml
# zerops.yaml
myapp:
  build:
    base: deno@1
    buildCommands:
      - deno cache main.ts
    deployFiles: ./
  run:
    start: deno run --allow-net --allow-env --allow-read main.ts
    ports:
      - port: 8000
        protocol: HTTP
```

### With deno.json Tasks
```json
{
  "tasks": {
    "start": "deno run --allow-net --allow-env --allow-read main.ts"
  }
}
```
```yaml
run:
  start: deno task start
```

## Gotchas
1. **Permissions are mandatory**: Without `--allow-net`, the app cannot open network ports — always set permissions
2. **Use `--allow-all` cautiously**: Grants all permissions — fine for Zerops (isolated container) but be explicit in production code
3. **`deno cache` for offline deps**: Cache dependencies during build so runtime doesn't need network for imports
4. **npm compat via `npm:` prefix**: Import npm packages with `import express from "npm:express"` — works out of the box

## See Also
- zerops://services/_common-runtime
- zerops://services/nodejs
- zerops://services/bun
- zerops://examples/zerops-yml-runtimes
