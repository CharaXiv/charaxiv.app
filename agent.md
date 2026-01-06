# CharaXiv Hono Agent Guide

This is a migration of CharaXiv from Go/templ to Hono/JSX for Cloudflare Workers compatibility.

## Migration Rationale

### Why Migrate?

The original Go/templ stack (Cloud Run + GCS) works well but ties us to GCP. Cloudflare Workers offers:
- Edge deployment (lower latency globally)
- Simpler pricing model
- D1 (SQLite at edge) + R2 (S3-compatible storage)

### Key Technical Decisions

#### CSS: Tailwind over OnceHandle

The original used templ's `OnceHandle` to inline component CSS once per request. Analysis showed:

| Metric | OnceHandle (inline) | Tailwind (external) |
|--------|---------------------|---------------------|
| First page load | 57.6 KB gzipped | ~59 KB (HTML + cached CSS) |
| Subsequent pages | 57.6 KB (styles re-sent) | **48.9 KB** (CSS cached) |
| HTMX partials | Include `<style>` blocks | Pure HTML (no CSS) |

**Decision**: Tailwind wins because:
1. CSS file cached after first load (~10KB gzipped)
2. HTMX partial responses are ~15% smaller (no inline styles)
3. Gzip compresses repetitive utility classes extremely well—the "class name bloat" concern is negligible after compression

#### JS: External Bundle over OnceHandle

The original inlined component scripts via `OnceHandle`. Issues:
- Scripts in HTMX partials don't re-execute reliably
- Risk of missing scripts if component first appears in a partial

**Decision**: Bundle all component JS into `/public/components.js`:
- ~4KB gzipped, cached forever
- Always available for HTMX partials
- Functions exist globally before any component renders

#### Framework: Hono + JSX

Hono was chosen because:
- API similar to chi (easy mental model transfer)
- Native JSX support (no React dependency)
- First-class Cloudflare Workers support
- Wrangler dev for local development

#### Storage: D1 + R2

The write coalescing pattern transfers directly:
- **D1**: SQLite write buffer (same as original)
- **R2**: Character JSON storage (replaces GCS)

**Write coalescing pattern** (implemented in `src/lib/storage/coalesce.ts`):
- Fast writes (~1ms): Buffer to D1 write_buffer table
- Reads: Load from R2, apply pending D1 writes, flush to R2, clear buffer
- This amortizes R2 write latency into page load

**AsyncStore interface** (`src/lib/storage/async-store.ts`):
- Unified async interface for character operations
- Two implementations:
  - `createMemoryStore()`: In-memory for Bun dev without D1/R2
  - `createD1R2Store(db, bucket)`: Production D1/R2 storage
- Route handlers access store via Hono context: `c.get('store')`
- Middleware auto-detects environment (D1/R2 bindings present → use D1/R2)

## Domain Model

### Character vs Sheet

A **Character** is the root entity representing a TRPG character. A character can have multiple **Sheets** (also called "systems") for different game systems or scenarios.

This enables players to:
- Use the same character across multiple game systems (CoC 6e, CoC 7e, Emoklore, etc.)
- Have separate sheets for different scenarios within the same system
- Share profile data (images, name) while keeping system-specific data separate

### Key Distinction: Memos

- **Character-level memos**: General character background, personality, etc. Shown on the character page.
- **Sheet-level memos**: Scenario-specific notes. Shown on the sheet page. Prefixed with "シナリオ" in UI.

### URL Structure

| Route | Page | Content |
|-------|------|----------|
| `/c/{characterId}` | Character page | Profile + character memos + list of sheets |
| `/c/{characterId}/{systemKey}/{sheetId}` | Sheet page | Profile + scenario memos + system-specific data |

### Legacy Reference

The `../charaxiv` directory contains the Go/templ implementation. Useful for:
- Understanding UI patterns and component structure
- Game system data schemas (CoC 6e, CoC 7e, Emoklore)
- Export formats (CCFOLIA, chat palettes)

## Project Structure

```
src/
  index.tsx        # Main Hono app with routes
  components/      # JSX components
    Layout.tsx
    Header.tsx
    Button.tsx
    NumberInput.tsx
    icons.tsx
  lib/
    types.ts       # TypeScript type definitions
    store.ts       # In-memory store (for local dev)
    storage/       # D1/R2 storage (for Workers deployment)
      coalesce.ts  # Write coalescing with D1 buffer + R2 JSON
      images.ts    # Image storage in R2
      index.ts     # Module exports
migrations/
  0001_init.sql    # D1 migration for write buffer
public/
  styles.css       # Tailwind output (generated)
  components.js    # Bundled component scripts
wrangler.toml      # Cloudflare Workers config
```

## Development

```bash
# Install dependencies
npm install

# Build Tailwind CSS
npm run css:build

# Run dev server with wrangler
npm run dev
```

Server runs on http://localhost:8002

## Migration Status

### Completed (P0, P1, P2, P3)
- [x] Project setup (Hono + Bun + Tailwind)
- [x] Layout and Header components
- [x] Icons (SVG components)
- [x] NumberInput with HTMX
- [x] Status panel (ability scores)
- [x] Parameters panel (HP/MP/SAN with ±1/±5 adjustment)
- [x] Random stats button (CoC 6e dice rules)
- [x] Skills panel (expandable rows, job/hobby/perm/temp inputs)
- [x] Multi-genre skills (運転, 芸術, ほかの言語)
- [x] Custom skills (add/delete/edit)
- [x] Grow checkbox toggle with OOB updates
- [x] Points display with live updates
- [x] Markdown editor (EasyMDE integration)
- [x] Markdown rendering in preview mode (marked.js)
- [x] Preview mode with OOB swaps
- [x] Profile auto-save (name/ruby)
- [x] Damage bonus input
- [x] Memo sections with auto-save
- [x] In-memory store with full CRUD
- [x] Image gallery (upload/delete/pin/navigation)

### Remaining (P3)
- [x] Image gallery component (upload/delete/pin/navigation)
- [x] Secret memo toggle (blur/reveal)
- [x] D1 integration for write buffer (schema + coalesce module)
- [x] R2 integration for character JSON (storage module)
- [x] Wrangler deployment config (migrations, bindings)
- [x] Switch routes to use D1/R2 storage (via AsyncStore interface)
- [ ] Gallery modal view (low priority)

## HTMX Patterns

### Core Patterns

- **OOB swaps** for updating multiple regions (status → skills → points)
- **hx-sync="queue last"** on number inputs to coalesce rapid changes
- **Optimistic UI** via `hx-on:before-request` for instant feedback

### OOB Swaps vs Chunk Replacement

Prefer **granular OOB swaps** over replacing a large HTML chunk when:
- Client-side state would be lost (e.g., `<details>` open/closed state, scroll position, focus)
- The alternative requires JavaScript to preserve/restore that state

Prefer **chunk replacement** when:
- Updates are self-contained with no client-side state to preserve
- The OOB target list would be excessively long or unpredictable

**Example**: Skill value changes use OOB to update input + breakdown + total + remaining points. This preserves the `<details>` expand/collapse state without client-side JavaScript.

## Deployment

Target: Cloudflare Workers via Wrangler

```bash
# Deploy
npm run deploy
```
