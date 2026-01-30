# .NET on Zerops

## Keywords
dotnet, csharp, aspnet, kestrel, nuget, blazor, entity framework, dotnet build, dotnet publish

## TL;DR
.NET on Zerops uses Kestrel as the default web server; publish as self-contained for portability or framework-dependent for smaller deploys.

## Zerops-Specific Behavior
- Versions: 6, 7, 8
- Base: Alpine (default)
- Web server: Kestrel (built-in)
- Working directory: `/var/www`
- No default port — Kestrel defaults to 5000 but configure explicitly
- Supports ASP.NET Core, Blazor, minimal APIs

## Configuration
```yaml
# zerops.yaml
myapp:
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

### Self-Contained Build
```yaml
build:
  buildCommands:
    - dotnet publish -c Release -r linux-musl-x64 --self-contained -o app
```

## Gotchas
1. **Set `ASPNETCORE_URLS`**: Must bind to `0.0.0.0` — Kestrel defaults to localhost which won't receive traffic
2. **Alpine = linux-musl-x64**: Use `linux-musl-x64` runtime identifier for self-contained builds on Alpine
3. **Cache NuGet packages**: `~/.nuget` cache avoids re-downloading packages every build
4. **Health checks**: ASP.NET `app.MapHealthChecks("/health")` works with Zerops readiness checks

## See Also
- zerops://services/_common-runtime
- zerops://examples/zerops-yml-runtimes
- zerops://platform/scaling
