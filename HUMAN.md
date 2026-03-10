# Human Actions Required

## After cloning this template for a new project

- [ ] Update Go module path in all `.go` files and `go.mod`
- [ ] Update `mise.toml` with app-specific `POSTGRES_DB` and `DATABASE_URL`
- [ ] Update `CLAUDE.md` with project-specific guidelines
- [ ] Run `npm install` to generate `package-lock.json`
- [ ] Run `just sync` to verify the codegen pipeline works end-to-end
- [ ] Install Go 1.25 via mise if not already available (`mise install go@1.25`)

## Dokku deployment setup

- [ ] Create Dokku app on server: `dokku apps:create <appname>`
- [ ] Create and link Postgres: `dokku postgres:create <appname>-db && dokku postgres:link <appname>-db <appname>`
- [ ] Set production config: `dokku config:set <appname> ENVIRONMENT=production COOKIE_SECURE=true COOKIE_DOMAIN=<domain>`
- [ ] Add git remote: `git remote add dokku dokku@<server>:<appname>`
- [ ] Configure DNS (Cloudflare A record pointing to server IP)
- [ ] Enable SSL: `dokku letsencrypt:enable <appname>`
- [ ] Set `DOKKU_HOST` in `mise.local.toml` for `just dokku-*` commands

## Database backups (if using Hetzner Object Storage)

- [ ] Create backup bucket and obtain access key / secret key
- [ ] Run `dokku postgres:backup-auth` with credentials
- [ ] Run `dokku postgres:backup-schedule` for daily backups
- [ ] Verify first backup completes: `dokku postgres:backup <appname>-db <bucket>`

## Monitoring (production)

- [ ] Expose Dokku postgres for Alloy monitoring: `dokku postgres:expose <appname>-db`
- [ ] Add `POSTGRES_DSN_<APPNAME>` to `/etc/alloy/alloy.env` on the server
- [ ] Add to Alloy config: `env("POSTGRES_DSN_<APPNAME>")` in data_source_names list
- [ ] Create Grafana dashboard in "Dokku Apps" folder (container metrics + nginx logs + app logs)
- [ ] Add Grafana synthetic HTTP uptime check (60s interval, alert on failure > 2m)

## OpenTelemetry (recommended)

OTel is built into the project template. Enable by setting env vars in Dokku:

- [ ] Set `OTEL_EXPORTER_OTLP_ENDPOINT` (e.g., `https://otlp-gateway-prod-au-southeast-1.grafana.net/otlp`)
- [ ] Set `OTEL_EXPORTER_OTLP_HEADERS` with Grafana Cloud credentials: `Authorization=Basic <base64(instanceId:apiKey)>`

This enables traces (Tempo), metrics (Prometheus), logs bridged to OTEL, and Go runtime metrics.
For local dev, set `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318` with the monitoring docker-compose profile.
