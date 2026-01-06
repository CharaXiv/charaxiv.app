# CharaXiv (Hono)

A TRPG character sheet management app built with Hono, HTMX, and JSX for Cloudflare Workers.

> **For AI agents:** Read `agent.md` and `CONTRACT.md` before starting development.

## Tech Stack

- **Hono** - Lightweight web framework for Cloudflare Workers
- **JSX** - Type-safe templating (no React)
- **HTMX** - Frontend interactivity without JavaScript frameworks
- **Tailwind CSS** - Utility-first CSS
- **Cloudflare D1** - SQLite database at the edge
- **Cloudflare R2** - S3-compatible object storage

## Development

### Prerequisites

- Node.js 18+
- npm

### Setup

```bash
# Install dependencies
npm install

# Build Tailwind CSS
npm run css:build

# Run D1 migrations (local)
npm run migrate

# Start dev server
npm run dev
```

Access the app at: **http://localhost:8002**

### Available Scripts

| Script | Description |
|--------|-------------|
| `npm run dev` | Start wrangler dev server |
| `npm run css` | Watch and rebuild Tailwind CSS |
| `npm run css:build` | Build minified Tailwind CSS |
| `npm run migrate` | Apply D1 migrations locally |
| `npm run migrate:remote` | Apply D1 migrations to production |
| `npm run deploy` | Deploy to Cloudflare Workers |

## Project Structure

```
src/
  index.tsx        # Main Hono app with routes
  components/      # JSX components
  lib/             # Shared utilities and types
    storage/       # D1/R2 storage layer
migrations/        # D1 database migrations
public/            # Static assets (CSS, JS)
wrangler.toml      # Cloudflare Workers config
```

## Architecture

### Storage

The app uses a **write coalescing** pattern:

```
Write path (~1ms):
  POST {path: "skills.回避.job", value: 5}
    → INSERT INTO write_buffer (D1)
    → 200 OK

Read path (page load):
  GET /character/{id}
    → Load JSON from R2
    → Apply pending writes from D1 buffer
    → Save updated JSON to R2 (flush)
    → Clear D1 buffer
    → Return response
```

**Why this model:**
- Fast ack to user (D1 write is ~1ms)
- Flush happens naturally on page load
- D1 buffer is durable (survives crashes)
- R2 writes are amortized into page load latency

### Environment Detection

The app auto-detects its environment:
- **With D1/R2 bindings**: Uses production storage
- **Without bindings**: Falls back to in-memory store

## Deployment

```bash
# Deploy to Cloudflare Workers
npm run deploy

# Apply migrations to production D1
npm run migrate:remote
```

## Related

- `../charaxiv` - Original Go/templ implementation (GCP)
