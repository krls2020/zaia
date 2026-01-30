# Init Commands & Migrations on Zerops

## Keywords
init, initCommands, prepareCommands, migration, migrate, zsc, execOnce, test tcp, seed, collectstatic, idempotent, startup, one-time, database migration

## TL;DR
`prepareCommands` = cached, runs once per change. `initCommands` = runs every container start. Use `zsc execOnce` for one-time initialization. Use `zsc test tcp` to wait for database readiness.

## Command Types

| Command | When it runs | Cached | Use for |
|---------|-------------|--------|---------|
| `prepareCommands` | First build or when changed | Yes | Package install, system deps |
| `buildCommands` | Every build | No | Compile, bundle, test |
| `initCommands` | Every container start | No | Migrations, cache warmup |
| `start` | After init completes | No | Application process |

## `zsc` Utility Commands

### `zsc execOnce` — One-Time Initialization

Runs a command only once per service lifetime. Subsequent container starts skip it.

```yaml
run:
  initCommands:
    - zsc execOnce createdb -- createdb myapp
    - zsc execOnce seed -- node seed.js
```

**Use cases:** Database creation, initial seed data, first-run setup

### `zsc test tcp` — Wait for Dependencies

Blocks until a TCP port is reachable. Use to wait for databases before running migrations.

```yaml
run:
  initCommands:
    - zsc test tcp db:5432              # wait for PostgreSQL
    - zsc test tcp cache:6379           # wait for Redis
    - php artisan migrate --force       # then run migration
```

## Migration Patterns by Framework

### Laravel (PHP)
```yaml
run:
  initCommands:
    - zsc test tcp db:5432
    - php artisan migrate --force
    - php artisan config:cache
    - php artisan route:cache
    - php artisan view:cache
```

### Django (Python)
```yaml
run:
  initCommands:
    - zsc test tcp db:5432
    - PYTHONPATH=./libs python manage.py migrate --noinput
    - PYTHONPATH=./libs python manage.py collectstatic --noinput
```

### Symfony (PHP)
```yaml
run:
  initCommands:
    - zsc test tcp db:5432
    - php bin/console doctrine:migrations:migrate --no-interaction
    - php bin/console cache:warmup
```

### Phoenix (Elixir)
Phoenix typically runs migrations during **build** (not init):
```yaml
build:
  buildCommands:
    - mix deps.get --only prod
    - MIX_ENV=prod mix ecto.create
    - MIX_ENV=prod mix ecto.migrate
    - MIX_ENV=prod mix compile
    - MIX_ENV=prod mix assets.deploy
    - MIX_ENV=prod mix phx.digest
    - MIX_ENV=prod mix release --overwrite
```

### NestJS / TypeORM (Node.js)
```yaml
run:
  initCommands:
    - zsc test tcp db:5432
    - npm run typeorm migration:run
```

### Payload CMS (Node.js)
Payload runs migrations automatically on startup — no explicit init needed.

## Static File Collection

Some frameworks need to collect/compile static assets at init time:

| Framework | Command | When |
|-----------|---------|------|
| Django | `python manage.py collectstatic --noinput` | initCommands |
| Symfony | `php bin/console asset-map:compile` | buildCommands |
| Phoenix | `mix phx.digest` | buildCommands |

## Idempotency Rules

All `initCommands` must be **idempotent** — safe to run multiple times:

| Framework | Safe command | Why |
|-----------|-------------|-----|
| Laravel | `artisan migrate --force` | Skips already-run migrations |
| Django | `manage.py migrate --noinput` | Tracks applied migrations |
| Symfony | `doctrine:migrations:migrate --no-interaction` | Migration versioning |
| Phoenix | `mix ecto.migrate` | Schema versioning |
| TypeORM | `typeorm migration:run` | Migration table tracking |

**Never put destructive commands in initCommands** — they run on every restart.

## Build-Time vs Runtime Migrations

| Timing | Pros | Cons | Use when |
|--------|------|------|----------|
| Build-time | Fails build if migration fails | Can't reach DB from build env | DB accessible from build |
| Runtime (initCommands) | DB always available | App starts slower | Default recommendation |

**Zerops recommendation:** Use `initCommands` with `zsc test tcp` for migrations.

## Gotchas
1. **initCommands run EVERY start**: Don't install packages here — use `prepareCommands`
2. **prepareCommands change = full rebuild**: Modifying prepare invalidates both cache layers
3. **Migrations must be idempotent**: Container restarts re-run all initCommands
4. **DB might not be ready**: Always use `zsc test tcp` before migration commands
5. **collectstatic in init, not build**: Django static files may need runtime env vars
6. **`zsc execOnce` is per-service**: Different services track independently

## See Also
- zerops://config/zerops-yml
- zerops://config/deploy-patterns
- zerops://services/_common-runtime
- zerops://gotchas/common
