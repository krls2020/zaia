# Elixir on Zerops

## Keywords
elixir, erlang, otp, phoenix, mix, release, beam, clustering, distributed, ecto

## TL;DR
Elixir on Zerops uses Mix releases for deployment; clustering across containers requires additional networking setup, and Erlang/OTP is pre-installed.

## Zerops-Specific Behavior
- Versions: 1.15, 1.16
- Base: Alpine (default), Erlang/OTP pre-installed
- Build tool: Mix (pre-installed)
- Working directory: `/var/www`
- No default port — must configure
- Deploy: Mix release (compiled BEAM bytecode)

## Configuration
```yaml
# zerops.yaml
myapp:
  build:
    base: elixir@1.16
    buildCommands:
      - mix deps.get --only prod
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
    envVariables:
      MIX_ENV: prod
      PHX_HOST: myapp.zerops.app
```

### Phoenix Framework
```yaml
build:
  buildCommands:
    - mix deps.get --only prod
    - MIX_ENV=prod mix assets.deploy
    - MIX_ENV=prod mix release
```

## Gotchas
1. **Clustering needs DNS setup**: BEAM clustering across containers requires `libcluster` with DNS strategy
2. **Cache `deps` and `_build`**: Elixir/Erlang compilation is slow — always cache
3. **Use Mix releases**: Don't deploy source code — compile to releases for production
4. **Set `MIX_ENV=prod`**: Both in build and runtime — affects compilation and behavior

## See Also
- zerops://services/_common-runtime
- zerops://services/gleam
- zerops://examples/zerops-yml-runtimes
