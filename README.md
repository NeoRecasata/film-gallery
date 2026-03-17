# Film Gallery

A self-hosted photo gallery built for film photographers. Organize your scans by roll, curate collections, and share your work through a minimal dark-themed gallery.

## Screenshots

**Public gallery** — Masonry grid with dark background, letting the photos speak for themselves.

![Public gallery with masonry grid layout](screenshots/gallery.png)

**Admin dashboard** — Overview of rolls, photos, collections, and storage usage.

![Admin dashboard](screenshots/admin-dashboard.png)

**Roll management** — Organize uploads by roll with shared metadata (camera, film stock, location).

![Rolls view with cover photos and metadata](screenshots/admin-rolls.png)

**Roll detail** — Edit roll metadata and manage individual photos. Per-photo visibility and metadata overrides.

![Roll detail with metadata form and photo grid](screenshots/admin-roll-detail.png)

## Features

- **Roll-based organization** — Upload photos by roll, just like you shoot. Each roll carries shared metadata (camera, film stock, lens, location, date) that individual photos inherit.
- **Masonry grid gallery** with fullscreen lightbox, keyboard/swipe navigation, and film metadata overlay
- **Collections** — Curate photos from any roll into themed sets
- **Per-photo overrides** — Override roll metadata on individual frames when needed
- **Visibility control** — Publish/draft rolls and hide individual photos within a roll
- **Automatic image processing** — Three WebP variants (400px, 1200px, 2400px) with BlurHash placeholders
- **Pluggable storage** — Local filesystem or any S3-compatible service (AWS, Cloudflare R2, MinIO)
- **Single-user** — One photographer per instance

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, chi, PostgreSQL, golang-migrate |
| Frontend | SvelteKit, Svelte 5, TypeScript, Tailwind CSS v4 |
| Image Processing | disintegration/imaging, chai2010/webp, go-blurhash |
| Storage | Local filesystem or S3-compatible |
| Auth | JWT (access + refresh tokens), bcrypt |

## Getting Started

### Prerequisites

- Go 1.22+ (with CGO enabled)
- Node.js 18+
- PostgreSQL 16 (or Docker)

### Development Setup

```bash
# Start PostgreSQL
docker run --rm -d --name gallery-db -p 5432:5432 \
  -e POSTGRES_USER=gallery \
  -e POSTGRES_PASSWORD=gallery \
  -e POSTGRES_DB=gallery \
  postgres:16-alpine

# Backend
cd backend
DATABASE_URL="postgres://gallery:gallery@localhost:5432/gallery?sslmode=disable" \
JWT_SECRET=dev-secret go run ./cmd/server
# → http://localhost:8080

# Frontend (separate terminal)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (proxies API to :8080)
```

Visit `http://localhost:5173/admin/setup` to create your admin account.

### Running Tests

```bash
cd backend && go test ./...
cd frontend && npm run check
```

## API

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/photos` | List visible photos (cursor pagination) |
| GET | `/api/photos/:slug` | Get photo with resolved metadata |
| GET | `/api/rolls` | List published rolls |
| GET | `/api/rolls/:slug` | Get roll with visible photos |
| GET | `/api/collections` | List collections |
| GET | `/api/collections/:slug` | Get collection with photos |
| GET | `/api/site` | Site settings |

### Auth

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/setup` | Create admin account (first-run only) |
| POST | `/api/auth/login` | Log in |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/logout` | Log out |
| POST | `/api/auth/change-password` | Change password |

### Admin (JWT required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/stats` | Dashboard stats |
| GET | `/api/admin/rolls` | List all rolls |
| GET | `/api/admin/rolls/:id` | Get roll with all photos |
| POST | `/api/admin/rolls` | Create roll |
| PATCH | `/api/admin/rolls/:id` | Update roll |
| DELETE | `/api/admin/rolls/:id` | Delete roll and photos |
| POST | `/api/admin/rolls/:id/photos` | Upload photos to roll |
| POST | `/api/admin/rolls/:id/photos/reorder` | Reorder photos |
| GET | `/api/admin/photos` | List all photos |
| PATCH | `/api/admin/photos/:id` | Update photo |
| DELETE | `/api/admin/photos/:id` | Delete photo |
| POST | `/api/admin/collections` | Create collection |
| PATCH | `/api/admin/collections/:id` | Update collection |
| DELETE | `/api/admin/collections/:id` | Delete collection |
| PUT | `/api/admin/collections/:id/photos` | Set collection photos |
| PATCH | `/api/admin/settings` | Update site settings |

## Configuration

All configuration via environment variables. See `backend/.env.example` for the full list.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `JWT_SECRET` | Yes | — | Secret for signing JWTs |
| `STORAGE_TYPE` | No | `local` | `local` or `s3` |
| `STORAGE_LOCAL_PATH` | No | `./data/photos` | Local storage path |
| `PORT` | No | `8080` | Backend port |

## Password Recovery

```bash
./server reset-password --email=your@email.com
```

## License

MIT
