# Film Gallery

Open-source, self-hostable photo gallery for film photographers. Upload, organize, and share your work through a minimal masonry grid gallery with a built-in admin dashboard.

**Status:** Backend complete. Frontend and Docker deployment in progress.

## Features

- Masonry grid gallery that respects each photo's natural aspect ratio
- Fullscreen lightbox with keyboard navigation and swipe support
- Collections for organizing photos into curated sets
- Film metadata fields (film stock, camera, lens) displayed alongside photos
- BlurHash placeholders for smooth lazy loading
- Admin dashboard with drag-and-drop uploads and batch processing
- Publish/draft workflow for controlling photo visibility
- Automatic image processing (thumbnail, medium, full-size WebP variants)
- Pluggable storage (local filesystem or S3-compatible)
- JWT authentication with secure refresh token rotation
- Self-hostable via Docker Compose with automatic TLS via Caddy
- Single-user design: one photographer per instance

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, chi, PostgreSQL, golang-migrate |
| Frontend | SvelteKit, TypeScript, Tailwind CSS v4 |
| Image Processing | disintegration/imaging, chai2010/webp, go-blurhash |
| Storage | Local filesystem, S3-compatible (AWS S3, Cloudflare R2, MinIO) |
| Auth | JWT (access + refresh tokens), bcrypt |
| Deployment | Docker Compose, Caddy |

## Project Structure

```
film-gallery/
├── backend/
│   ├── cmd/server/          # Entry point, CLI commands
│   ├── internal/
│   │   ├── api/             # HTTP handlers and router
│   │   ├── auth/            # JWT service and middleware
│   │   ├── config/          # Environment-based configuration
│   │   ├── db/              # Database connection and migrations
│   │   ├── media/           # Image processing (resize, WebP, BlurHash)
│   │   ├── models/          # Data types
│   │   ├── slug/            # URL-friendly slug generation
│   │   └── storage/         # Storage interface (local, S3)
│   └── migrations/          # PostgreSQL schema migrations
├── frontend/                # SvelteKit app (coming soon)
├── docker-compose.yml       # Production deployment (coming soon)
└── Caddyfile                # Reverse proxy config (coming soon)
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL 16
- Docker (for running Postgres locally)

### Development Setup

```bash
# Clone the repo
git clone https://github.com/NeoRecasata/film-gallery.git
cd film-gallery

# Start PostgreSQL
docker run --rm -d --name gallery-db -p 5432:5432 \
  -e POSTGRES_USER=gallery \
  -e POSTGRES_PASSWORD=gallery \
  -e POSTGRES_DB=gallery \
  postgres:16-alpine

# Configure and run the backend
cd backend
cp .env.example .env
# Edit .env if needed (defaults work with the Docker Postgres above)

go run ./cmd/server
# Server starts on http://localhost:8080
```

### Verify it works

```bash
curl http://localhost:8080/api/health
# {"status":"ok"}
```

### Running tests

```bash
cd backend
go test ./...
```

## API

### Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/photos` | List published photos (cursor pagination, filterable) |
| GET | `/api/photos/:slug` | Get a single photo |
| GET | `/api/collections` | List all collections |
| GET | `/api/collections/:slug` | Get collection with photos |
| GET | `/api/site` | Get site settings |

### Auth Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/setup` | Create initial admin account (first-run only) |
| POST | `/api/auth/login` | Log in, returns JWT |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/logout` | Clear refresh token |
| POST | `/api/auth/change-password` | Change password (authenticated) |

### Admin Endpoints (JWT required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/admin/photos` | Upload photo (multipart) |
| PATCH | `/api/admin/photos/:id` | Update photo metadata |
| DELETE | `/api/admin/photos/:id` | Delete photo |
| POST | `/api/admin/photos/reorder` | Batch reorder photos |
| POST | `/api/admin/collections` | Create collection |
| PATCH | `/api/admin/collections/:id` | Update collection |
| DELETE | `/api/admin/collections/:id` | Delete collection |
| PUT | `/api/admin/collections/:id/photos` | Set collection photos |
| PUT | `/api/admin/site` | Update site settings |

## Configuration

All configuration is via environment variables. See `backend/.env.example` for all options.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `JWT_SECRET` | Yes | - | Secret for signing JWTs |
| `STORAGE_TYPE` | No | `local` | `local` or `s3` |
| `STORAGE_LOCAL_PATH` | No | `./data/photos` | Path for local photo storage |
| `PORT` | No | `8080` | Server port |

For S3 configuration, see `.env.example`.

## Password Recovery

If you lose your admin credentials:

```bash
./server reset-password --email=your@email.com
```

Or with Docker:

```bash
docker compose exec backend ./server reset-password --email=your@email.com
```

## Roadmap

- [x] Go REST API backend
- [ ] SvelteKit frontend (public gallery + admin dashboard)
- [ ] Docker Compose deployment with Caddy
- [ ] EXIF auto-extraction
- [ ] Single-binary mode via Go embed

## License

MIT
