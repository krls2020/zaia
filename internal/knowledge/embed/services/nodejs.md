# Node.js on Zerops

## Keywords
nodejs, node, javascript, typescript, npm, yarn, pnpm, express, nextjs, nuxt, nest, bun alternative

## TL;DR
Node.js on Zerops supports versions 18-22, uses Alpine base by default, and requires explicit `node_modules` in `deployFiles` if not bundling.

## Zerops-Specific Behavior
- Versions: 18, 20, 22
- Base: Alpine (default)
- Package managers: npm, yarn, pnpm (pre-installed)
- Working directory: `/var/www`
- No default port — must configure in `zerops.yaml`
- `start` command required in `run` section

## Configuration
```yaml
# zerops.yaml
myapp:
  build:
    base: nodejs@22
    buildCommands:
      - npm ci
      - npm run build
    deployFiles:
      - dist
      - node_modules
      - package.json
    cache:
      - node_modules
  run:
    start: node dist/index.js
    ports:
      - port: 3000
        protocol: HTTP
```

## Framework Patterns

### Next.js
```yaml
build:
  buildCommands:
    - npm ci
    - npm run build
  deployFiles:
    - .next
    - node_modules
    - package.json
    - next.config.js
run:
  start: npm start
```

### NestJS
```yaml
build:
  buildCommands:
    - npm ci
    - npm run build
  deployFiles:
    - dist
    - node_modules
    - package.json
run:
  start: node dist/main.js
```

## SSR vs SSG Deployment

| Mode | Build base | Run base | Deploy |
|------|-----------|----------|--------|
| SSR | `nodejs@22` | `nodejs@22` | `.next`, `node_modules`, `package.json` |
| SSG | `nodejs@22` | `static` | `out/~` or `dist/~` (tilde = contents) |

**SSR** = full Node.js runtime needed. **SSG** = build produces static HTML, deploy to `static` service.

## Framework Adapters

Some frameworks need explicit adapters for Node.js SSR on Zerops:

| Framework | Adapter needed | Notes |
|-----------|---------------|-------|
| Next.js | None (built-in) | SSR default, `output: 'export'` for SSG |
| Nuxt | None (built-in) | `nuxi build` for SSR, `nuxi generate` for SSG |
| Astro | `@astrojs/node` | Required for SSR mode |
| SvelteKit | `@sveltejs/adapter-node` | Required for Node.js target |
| Qwik | `express` adapter | Needs express server |
| Remix | `@remix-run/express` | Express adapter for Node.js |

## Package Manager

**pnpm** is the preferred package manager on Zerops (pre-installed, faster, less disk):
```yaml
buildCommands:
  - pnpm i && pnpm build
cache:
  - node_modules
  - .pnpm-store
```

## Environment Variables

| Variable | Value | Required |
|----------|-------|----------|
| `HOST` | `0.0.0.0` | Yes — Node.js must listen on all interfaces |
| `PORT` | `3000` | Convention (match zerops.yml port) |
| `NODE_ENV` | `production` | Recommended |

## Gotchas
1. **Include `node_modules` in deployFiles**: Unless using a bundler that inlines all deps — runtime has no `npm install`
2. **No default port**: Must explicitly set port in `zerops.yaml` — Node.js doesn't auto-detect
3. **Use `npm ci` not `npm install`**: Ensures reproducible builds from lockfile
4. **Cache `node_modules`**: Speeds up builds significantly — add to `build.cache`
5. **`HOST=0.0.0.0` required**: Without it, app may bind to localhost only — unreachable from Zerops routing
6. **SSG needs tilde syntax**: Deploy `out/~` not `out` to avoid nested directory on static service

## See Also
- zerops://services/_common-runtime
- zerops://services/bun
- zerops://services/static
- zerops://config/deploy-patterns
- zerops://examples/zerops-yml-runtimes
- zerops://examples/connection-strings
