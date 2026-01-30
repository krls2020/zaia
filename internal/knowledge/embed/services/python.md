# Python on Zerops

## Keywords
python, pip, flask, fastapi, django, uvicorn, gunicorn, virtualenv, requirements.txt, poetry

## TL;DR
Python on Zerops uses Alpine base with pip; system packages installed via `prepareCommands` persist across builds (unlike `buildCommands`), and there's no default port.

## Zerops-Specific Behavior
- Versions: 3.9, 3.10, 3.11, 3.12
- Base: Alpine (default)
- Package manager: pip (pre-installed)
- Working directory: `/var/www`
- No default port — must configure
- Use `addToRunPrepare` for runtime system dependencies

## Configuration
```yaml
# zerops.yaml
myapp:
  build:
    base: python@3.12
    buildCommands:
      - pip install -r requirements.txt --target=./libs
    deployFiles:
      - ./
    cache:
      - libs
  run:
    start: python -m uvicorn main:app --host 0.0.0.0 --port 8000
    ports:
      - port: 8000
        protocol: HTTP
```

## Framework Patterns

### FastAPI
```yaml
build:
  buildCommands:
    - pip install -r requirements.txt --target=./libs
  deployFiles: ./
run:
  start: PYTHONPATH=./libs python -m uvicorn main:app --host 0.0.0.0 --port 8000
```

### Django
```yaml
build:
  buildCommands:
    - pip install -r requirements.txt --target=./libs
    - PYTHONPATH=./libs python manage.py collectstatic --noinput
  deployFiles: ./
run:
  start: PYTHONPATH=./libs gunicorn myproject.wsgi:application --bind 0.0.0.0:8000
```

### Flask
```yaml
run:
  start: PYTHONPATH=./libs gunicorn app:app --bind 0.0.0.0:5000
```

## Gotchas
1. **pip installs to system Python**: Use `--target=./libs` and set `PYTHONPATH` to keep deps portable
2. **Alpine musl issues**: Some C extension packages (numpy, pandas) may need `prepareCommands` to install build tools
3. **No default port**: Must explicitly bind to `0.0.0.0:PORT` — localhost won't work
4. **`addToRunPrepare` for runtime deps**: System libraries needed at runtime (e.g., libpq) go in `addToRunPrepare`, not `prepareCommands`

## See Also
- zerops://services/_common-runtime
- zerops://examples/zerops-yml-runtimes
- zerops://examples/connection-strings
