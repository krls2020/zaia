# zerops.yml Runtime Examples

## Keywords
examples, zerops.yml, nodejs example, python example, php example, go example, java example, rust example, multi-service, monorepo

## TL;DR
Copy-paste ready zerops.yml configurations for all supported runtimes and common framework patterns.

## Node.js (Express)
```yaml
api:
  build:
    base: nodejs@22
    buildCommands:
      - npm ci
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

## Node.js (Next.js)
```yaml
web:
  build:
    base: nodejs@22
    buildCommands:
      - npm ci
      - npm run build
    deployFiles:
      - .next
      - node_modules
      - package.json
      - next.config.js
      - public
    cache:
      - node_modules
      - .next/cache
  run:
    start: npm start
    ports:
      - port: 3000
        protocol: HTTP
```

## Python (FastAPI)
```yaml
api:
  build:
    base: python@3.12
    buildCommands:
      - pip install -r requirements.txt --target=./libs
    deployFiles: ./
    cache:
      - libs
  run:
    start: PYTHONPATH=./libs python -m uvicorn main:app --host 0.0.0.0 --port 8000
    ports:
      - port: 8000
        protocol: HTTP
```

## Python (Django)
```yaml
web:
  build:
    base: python@3.12
    buildCommands:
      - pip install -r requirements.txt --target=./libs
      - PYTHONPATH=./libs python manage.py collectstatic --noinput
    deployFiles: ./
    cache:
      - libs
  run:
    start: PYTHONPATH=./libs gunicorn myproject.wsgi:application --bind 0.0.0.0:8000
    ports:
      - port: 8000
        protocol: HTTP
```

## PHP (Laravel)
```yaml
web:
  build:
    base: php-nginx@8.3
    prepareCommands:
      - apk add --no-cache icu-dev
      - docker-php-ext-install intl pdo_pgsql
    buildCommands:
      - composer install --no-dev --optimize-autoloader
      - php artisan config:cache
      - php artisan route:cache
    deployFiles: ./
    cache:
      - vendor
  run:
    documentRoot: public
    ports:
      - port: 80
        protocol: HTTP
    envVariables:
      APP_ENV: production
```

## Go
```yaml
api:
  build:
    base: go@1.22
    buildCommands:
      - go build -o app ./cmd/server
    deployFiles:
      - app
  run:
    start: ./app
    ports:
      - port: 8080
        protocol: HTTP
```

## Rust
```yaml
api:
  build:
    base: rust@latest
    buildCommands:
      - cargo build --release
    deployFiles:
      - target/release/myapp
    cache:
      - target
      - ~/.cargo/registry
  run:
    start: ./myapp
    ports:
      - port: 8080
        protocol: HTTP
```

## Java (Spring Boot)
```yaml
api:
  build:
    base: java@21
    buildCommands:
      - ./mvnw package -DskipTests
    deployFiles:
      - target/app.jar
    cache:
      - .m2
  run:
    start: java -Xmx512m -jar app.jar
    ports:
      - port: 8080
        protocol: HTTP
```

## .NET
```yaml
api:
  build:
    base: dotnet@8
    buildCommands:
      - dotnet publish -c Release -o app
    deployFiles:
      - app
    cache:
      - ~/.nuget
  run:
    start: ./app/MyApp
    ports:
      - port: 5000
        protocol: HTTP
    envVariables:
      ASPNETCORE_URLS: http://0.0.0.0:5000
```

## Elixir (Phoenix)
```yaml
web:
  build:
    base: elixir@1.16
    buildCommands:
      - mix deps.get --only prod
      - MIX_ENV=prod mix assets.deploy
      - MIX_ENV=prod mix release
    deployFiles:
      - _build/prod/rel/myapp
    cache:
      - deps
      - _build
  run:
    start: _build/prod/rel/myapp/bin/myapp start
    ports:
      - port: 4000
        protocol: HTTP
```

## Static (React/Vue/Angular SPA)
```yaml
web:
  build:
    base: nodejs@22
    buildCommands:
      - npm ci
      - npm run build
    deployFiles:
      - dist
    cache:
      - node_modules
  run:
    documentRoot: dist
```

## Multi-Service Monorepo
```yaml
api:
  build:
    base: nodejs@22
    buildCommands:
      - npm ci
      - npm run build:api
    deployFiles:
      - packages/api/dist
      - node_modules
    cache:
      - node_modules
  run:
    start: node packages/api/dist/index.js
    ports:
      - port: 3000
        protocol: HTTP

web:
  build:
    base: nodejs@22
    buildCommands:
      - npm ci
      - npm run build:web
    deployFiles:
      - packages/web/dist
    cache:
      - node_modules
  run:
    documentRoot: packages/web/dist
```

## Gleam (Wisp + Mist)
```yaml
api:
  build:
    base: gleam@1
    buildCommands:
      - gleam export erlang-shipment
    deployFiles:
      - build/erlang-shipment
    cache:
      - build
  run:
    start: ./build/erlang-shipment/entrypoint.sh run
    ports:
      - port: 3000
        protocol: HTTP
```

## Bun
```yaml
api:
  build:
    base: bun@1
    buildCommands:
      - bun install
    deployFiles:
      - node_modules
      - src
      - package.json
    cache:
      - node_modules
  run:
    start: bun run src/index.ts
    ports:
      - port: 3000
        protocol: HTTP
```

## Deno
```yaml
api:
  build:
    base: deno@2
    buildCommands:
      - deno cache src/main.ts
    deployFiles: ./
  run:
    start: deno run --allow-net --allow-env src/main.ts
    ports:
      - port: 8000
        protocol: HTTP
```

## Discord Bot (No HTTP — Background Process)
```yaml
bot:
  build:
    base: nodejs@22
    buildCommands:
      - npm ci
      - npm run build
    deployFiles:
      - dist
      - node_modules
      - package.json
  run:
    start: node dist/bot.js
```
Note: No `ports` section — background process without HTTP server.

## Multi-Base: Elixir → Alpine (Phoenix Release)
```yaml
web:
  build:
    base: elixir@1.16
    buildCommands:
      - mix local.hex --force && mix local.rebar --force
      - mix deps.get --only prod
      - MIX_ENV=prod mix compile
      - MIX_ENV=prod mix assets.deploy
      - MIX_ENV=prod mix phx.digest
      - MIX_ENV=prod mix release --overwrite
    deployFiles:
      - _build/prod/rel/myapp
    cache:
      - deps
      - _build
  run:
    base: alpine@latest
    start: _build/prod/rel/myapp/bin/myapp start
    ports:
      - port: 4000
        protocol: HTTP
    envVariables:
      PHX_SERVER: "true"
```

## Multi-Base: Node.js → Static (SSG)
```yaml
web:
  build:
    base: nodejs@22
    buildCommands:
      - pnpm i && pnpm build
    deployFiles:
      - dist/~
    cache:
      - node_modules
  run:
    base: static
```

## See Also
- zerops://config/zerops-yml
- zerops://config/deploy-patterns
- zerops://examples/connection-strings
- zerops://services/_common-runtime
