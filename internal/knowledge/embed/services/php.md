# PHP on Zerops

## Keywords
php, laravel, wordpress, composer, nginx, apache, php-fpm, php-nginx, php-apache, document root

## TL;DR
PHP on Zerops comes in two variants: `php-nginx` (recommended) and `php-apache`, uses Composer for deps, and serves from `/var/www` with configurable document root.

## Zerops-Specific Behavior
- Versions: 8.1, 8.2, 8.3, 8.4
- Variants: `php-nginx` (recommended), `php-apache`
- Base: Alpine (default)
- Package manager: Composer (pre-installed)
- Document root: configurable (default `/var/www`)
- Xdebug: Available, configure via env vars

## Configuration
```yaml
# zerops.yaml
myapp:
  build:
    base: php-nginx@8.3
    buildCommands:
      - composer install --no-dev --optimize-autoloader
    deployFiles: ./
    cache:
      - vendor
  run:
    documentRoot: public
    ports:
      - port: 80
        protocol: HTTP
```

## Framework Patterns

### Laravel
```yaml
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
  envVariables:
    APP_ENV: production
```

### WordPress
```yaml
build:
  base: php-apache@8.3
  buildCommands:
    - composer install --no-dev
  deployFiles: ./
run:
  documentRoot: ""
```

## Apache vs Nginx Variants

| Variant | Runtime type | Use when |
|---------|-------------|----------|
| `php-nginx@8.3` | Nginx + PHP-FPM | Modern frameworks (Laravel, Symfony) — **recommended** |
| `php-apache@8.3` | Apache + mod_php | Apps requiring `.htaccess` (WordPress, legacy PHP) |

**Build base is always generic `php@8.3`** (for Composer) — the variant only matters for run:

```yaml
# Build uses generic PHP
build:
  base: php@8.3
  buildCommands:
    - composer install --no-dev --optimize-autoloader

# Run uses variant
run:
  base: php-nginx@8.3
  documentRoot: public
```

## Custom Nginx Config

Use `siteConfigPath` to provide custom Nginx configuration:

```yaml
run:
  siteConfigPath: nginx.conf
  documentRoot: public
```

## Trusted Proxies

Zerops routes traffic through a reverse proxy. Frameworks must trust it:

```yaml
# Laravel
envVariables:
  TRUSTED_PROXIES: "127.0.0.1,10.0.0.0/8"

# Symfony
envVariables:
  TRUSTED_PROXIES: "127.0.0.1,10.0.0.0/8"
```

Without this, CSRF validation, HTTPS detection, and client IP resolution break.

## Logging

Use syslog for structured logging (Zerops captures syslog output):

```php
// Laravel: config/logging.php — use 'syslog' channel
// Symfony: monolog.yaml — use SyslogHandler
```

## Sessions

**File-based sessions break with multiple containers.** Use Redis/Valkey:

```yaml
envVariables:
  SESSION_DRIVER: redis
  REDIS_HOST: cache
  REDIS_PORT: "6379"
envSecrets:
  REDIS_PASSWORD: ${cache_password}
```

## Gotchas
1. **Document root matters**: Laravel needs `public`, WordPress uses root — misconfigured doc root = 404
2. **Composer cache in vendor/**: Add `vendor` to `build.cache` for faster builds
3. **PHP extensions need `prepareCommands`**: Install with `docker-php-ext-install` in prepareCommands, not buildCommands
4. **php-nginx vs php-apache**: Use `php-nginx` unless your app specifically requires Apache (.htaccess)
5. **Trusted proxies required**: Without proxy config, CSRF breaks behind Zerops L7 balancer
6. **File sessions don't scale**: Multiple containers don't share filesystem — use Redis for sessions
7. **Build base ≠ run base**: Build uses generic `php@8.3`, run uses `php-nginx@8.3` or `php-apache@8.3`

## See Also
- zerops://services/_common-runtime
- zerops://services/nginx
- zerops://operations/production-checklist
- zerops://operations/init-commands
- zerops://examples/zerops-yml-runtimes
