# Bun on Zerops

## Keywords
bun, javascript, typescript, bun runtime, bunx, bun.lockb, fast js, bun install

## TL;DR
Bun on Zerops is a drop-in Node.js replacement with native TypeScript support; uses `bun.lockb` lockfile and `bunx` instead of `npx`.

## Zerops-Specific Behavior
- Versions: 1.1+
- Base: Alpine (default)
- Package manager: `bun install` (npm-compatible)
- Lockfile: `bun.lockb` (binary format)
- Working directory: `/var/www`
- No default port — must configure
- npx replacement: `bunx`

## Configuration
```yaml
# zerops.yaml
myapp:
  build:
    base: bun@1.1
    buildCommands:
      - bun install --production
      - bun run build
    deployFiles:
      - dist
      - node_modules
      - package.json
    cache:
      - node_modules
  run:
    start: bun run dist/index.ts
    ports:
      - port: 3000
        protocol: HTTP
```

## Gotchas
1. **npm-compatible but not identical**: Most npm packages work, but some with native Node.js APIs may not
2. **`bun.lockb` is binary**: Cannot be manually edited or diffed — regenerate with `bun install`
3. **Include `node_modules` in deployFiles**: Same as Node.js — runtime doesn't have `bun install`

## See Also
- zerops://services/_common-runtime
- zerops://services/nodejs
- zerops://services/deno
- zerops://examples/zerops-yml-runtimes
