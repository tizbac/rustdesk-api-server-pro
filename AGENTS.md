# AGENTS.md

This repository contains a RustDesk API server implementation with Go backend and Vue frontend.

## Key Commands

- `make build` - Build the entire application (Go backend + Vue frontend)
- `cd backend && go build` - Build just the Go backend
- `cd soybean-admin && pnpm build` - Build just the Vue frontend
- `./build/rustdesk-api-server-pro sync` - Synchronize database schema
- `./build/rustdesk-api-server-pro start` - Start the server
- `./build/rustdesk-api-server-pro user add admin password --admin` - Add admin user

## Architecture

- Go backend in `/backend/` directory
- Vue frontend in `/soybean-admin/` directory
- Docker image available at `ghcr.io/lantongxue/rustdesk-api-server-pro:latest`
- Default server port: 8080
- Default database: SQLite (default `server.db` file)

## Deployment

- For Docker deployment, use `docker run` or `docker-compose.yaml`
- Environment variables: ADMIN_USER, ADMIN_PASS, TZ
- Configuration file path: `/app/data/server.yaml` (default)
- Reverse proxy required for web UI access

## Configuration

- The web client URL can be configured in `server.yaml` under `webClient.url` (default: `https://rustdesk.etitech.net/webclient`)

## Development

- Go >= 1.21.4 required
- Node.js and pnpm required for frontend
- Build process combines Go backend and Vue frontend into single binary
- Configuration is done via server.yaml file