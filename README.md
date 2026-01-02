# CharaXiv

A TRPG character sheet management app built with Go, HTMX, and templ.

## Tech Stack

- **Go** - Backend server
- **templ** - Type-safe HTML templating
- **HTMX** - Frontend interactivity without JavaScript frameworks
- **sqlc** - Type-safe SQL (pending)
- **SQLite** - Database (pending)

## Project Structure

```
charaxiv/
├── cmd/dev/           # Development server with live reload
├── templates/         # templ templates
│   ├── styles.templ   # Design system (CSS variables)
│   ├── layout.templ   # Base HTML layout
│   └── character.templ # Character sheet components
├── main.go            # Application entry point
├── nginx.conf         # Proxy config for WebSocket support
├── .air.*.toml        # Hot reload configuration
└── sqlc.yaml          # SQL code generation config
```

## Development

### Prerequisites

- Go 1.23+
- templ CLI: `go install github.com/a-h/templ/cmd/templ@latest`
- air CLI: `go install github.com/air-verse/air@latest`
- sqlc CLI: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

### Running the Dev Server

```bash
# Start nginx proxy (first time or after config changes)
nginx -c /home/exedev/charaxiv/nginx.conf

# Start dev server with hot reload
./bin/dev
```

Access the app at: **https://charaxiv.exe.xyz:8080/**

### How Live Reload Works

1. Edit a `.go` or `.templ` file → air runs `gofmt` and `templ fmt`
2. Edit a `.templ` file → air runs `templ generate`
3. Generated `_templ.go` changes → air rebuilds and restarts server
4. Server starts → triggers `/reload` endpoint
5. Reloader waits for `/health` to return 200
6. Reloader sends WebSocket message to all connected browsers
7. Browsers automatically refresh

### Auto-formatting

Files are automatically formatted on save:
- **Go files**: `gofmt -w`
- **Templ files**: `templ fmt`

### Building for Production

```bash
templ generate
go build -o charaxiv .
./charaxiv
```

## Design System

All styles use CSS custom properties defined in `templates/styles.templ`.

### Colors

```css
--color-primary         /* #2563eb - Main brand color */
--color-primary-dark    /* #1d4ed8 - Darker variant */
--color-primary-light   /* #dbeafe - Light variant */
--color-background      /* #f5f5f5 - Page background */
--color-surface         /* #ffffff - Card/panel background */
--color-text            /* #333333 - Primary text */
--color-text-muted      /* #64748b - Secondary text */
--color-border          /* #e2e8f0 - Borders */
--color-border-light    /* #f1f5f9 - Subtle borders */
```

### Spacing (4px grid)

```css
--space-1: 4px
--space-2: 8px
--space-3: 12px
--space-4: 16px
--space-5: 20px
--space-6: 24px
--space-8: 32px
```

### Typography

```css
--font-size-xs: 0.75rem
--font-size-sm: 0.875rem
--font-size-base: 1rem
--font-size-lg: 1.125rem
--font-size-xl: 1.25rem
--font-size-2xl: 1.5rem

--font-weight-normal: 400
--font-weight-medium: 500
--font-weight-semibold: 600
--font-weight-bold: 700
```

### Other Tokens

```css
--radius-sm: 4px
--radius-md: 8px
--radius-lg: 12px

--shadow-sm: 0 1px 3px rgba(0,0,0,0.1)
--shadow-md: 0 2px 8px rgba(0,0,0,0.1)

--transition-fast: 150ms ease
--transition-normal: 250ms ease
```

## Responsive Layout

The character sheet uses a responsive two-column layout:

| Breakpoint | Main Layout | Right Column |
|------------|-------------|--------------|
| < 640px | Single column (max 440px) | Single column |
| 640px+ | 1fr / 320px | Single column |
| 768px+ | 1fr / 320px, 8px gap, max 768px | Single column |
| 1024px+ | 1fr / 648px, 16px gap, max 1104px | 1fr / 1fr |
| 1536px+ | 440px / 968px, 16px gap, max 1536px | 1fr / 2fr |

## Gotchas

### Compression Middleware + templ

Chi's `Compress` middleware requires `Content-Type` to be set **before** writing the response body. templ doesn't set this automatically since it writes to a generic `io.Writer`.

```go
// ❌ Won't compress - Content-Type not set when Write() is called
r.Get("/", func(w http.ResponseWriter, r *http.Request) {
    templates.MyPage().Render(r.Context(), w)
})

// ✅ Will compress - Content-Type set before render
r.Get("/", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    templates.MyPage().Render(r.Context(), w)
})
```

**Why?** The middleware checks `Content-Type` on the first `Write()` call to decide whether to compress. Go's auto-detection happens *during* the first write, which is too late.

## Component Patterns

### Inlined Styles with OnceHandle

Styles are co-located with components but only rendered once:

```go
var myStyles = templ.NewOnceHandle()

templ MyComponent() {
    @myStyles.Once() {
        <style>
            .my-class { ... }
        </style>
    }
    <div class="my-class">...</div>
}
```

### HTMX Partial Updates

Use `hx-*` attributes for server-driven interactivity:

```html
<button 
    hx-post="/api/increment"
    hx-target="#counter"
    hx-swap="outerHTML">
    +1
</button>
```

### Out-of-Band (OOB) Swaps

Update multiple elements from a single response:

```go
// Primary response (swaps into hx-target)
templates.PrimaryComponent().Render(ctx, w)

// OOB response (swaps by ID anywhere on page)
templates.DependentComponentOOB().Render(ctx, w)
```

```html
<!-- OOB component template -->
<div id="dependent" hx-swap-oob="true">...</div>
```

## Git Workflow

Make small, focused commits:

```bash
git add <specific-files>
git commit -m "Short description

- Detail 1
- Detail 2"
```
