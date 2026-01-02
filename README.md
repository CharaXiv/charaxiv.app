# CharaXiv

A TRPG character sheet management app built with Go, HTMX, and templ.

## Tech Stack

- **Go** - Backend server
- **templ** - Type-safe HTML templating
- **HTMX** - Frontend interactivity without JavaScript frameworks
- **SQLite** - Database (with sqlc for type-safe queries)
- **GCS** - Google Cloud Storage for images and assets
- **Terraform** - Infrastructure as Code

## Project Structure

```
charaxiv/
├── cmd/dev/           # Development server with live reload
├── templates/         # templ templates
│   ├── styles.templ   # Design system (CSS variables)
│   ├── layout.templ   # Base HTML layout
│   └── character.templ # Character sheet components
├── storage/           # Storage abstraction layer
│   ├── storage.go     # Storage interface
│   ├── gcs.go         # Google Cloud Storage implementation
│   └── memory.go      # In-memory implementation (for testing)
├── terraform/         # Infrastructure as Code
│   ├── modules/       # Reusable Terraform modules
│   │   ├── gcs/       # GCS bucket + service account
│   │   └── cloudrun/  # Cloud Run deployment
│   ├── environments/  # Environment-specific configs
│   │   ├── dev/       # Development (US-West)
│   │   └── prd/       # Production (Tokyo)
│   ├── bootstrap.yaml # Deployment Manager for tfstate bucket
│   └── bootstrap.jinja
├── main.go            # Application entry point
├── Dockerfile         # Container build
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

**The dev server handles everything automatically.** Just edit your files and the browser will refresh with your changes. No need to manually run `templ generate` or `go build`.

### How Live Reload Works

The dev server (`./bin/dev`) runs multiple processes via [air](https://github.com/air-verse/air):

1. **Formatter** - Watches `.go` and `.templ` files, runs `gofmt` and `templ fmt` on changes
2. **Templ Generator** - Watches `.templ` files, runs `templ generate` on changes
3. **Server Builder** - Watches `.go` files (including generated `_templ.go`), rebuilds and restarts the server
4. **Reloader** - WebSocket server that notifies browsers to refresh

The reload flow:
1. Edit a `.go` or `.templ` file
2. Formatter auto-formats the file
3. If `.templ` file, generator creates/updates `_templ.go`
4. Server rebuilds and restarts
5. Reloader waits for `/health` endpoint to return 200
6. Reloader sends WebSocket message to connected browsers
7. Browsers automatically refresh

### Auto-formatting

Files are automatically formatted when saved (handled by the dev server):
- **Go files**: `gofmt -w`
- **Templ files**: `templ fmt`

### Building for Production

```bash
templ generate
go build -o charaxiv .
./charaxiv
```

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `GCS_BUCKET` | GCS bucket name for storage | No (falls back to in-memory) |
| `PORT` | Server port | No (default: 8000) |
| `DEV` | Enable dev mode (live reload trigger) | No |

## Infrastructure

### Architecture

```
Development (exe.dev VM - US-West):
  App Server → GCS (charaxiv-app-dev-assets)

Production (Cloud Run - Tokyo):
  Users → Cloud Run → GCS (charaxiv-app-prd-assets)
                    → SQLite (local + Litestream backup)
```

### Storage

The app uses a `Storage` interface that abstracts away the storage backend:

```go
type Storage interface {
    Upload(ctx, key, data, contentType) error
    Download(ctx, key) (io.ReadCloser, error)
    Delete(ctx, key) error
    Exists(ctx, key) (bool, error)
    SignedURL(ctx, key, expiry) (string, error)
    SignedUploadURL(ctx, key, contentType, expiry) (string, error)
    PublicURL(key) string
}
```

Implementations:
- **GCS** - Production storage, uses Application Default Credentials
- **Memory** - In-memory storage for testing (no external dependencies)

### Terraform Setup

**Prerequisites:**
- Google Cloud SDK (`gcloud`)
- Terraform
- A GCP project

**1. Authenticate with GCP:**
```bash
gcloud auth login
gcloud auth application-default login
gcloud config set project YOUR_PROJECT_ID
```

**2. Bootstrap tfstate bucket (one-time):**
```bash
gcloud deployment-manager deployments create tfstate \
  --config terraform/bootstrap.yaml
```

**3. Initialize and apply Terraform:**
```bash
cd terraform/environments/dev
cp backend.conf.example backend.conf
cp terraform.tfvars.example terraform.tfvars
# Edit both files with your values

terraform init -backend-config=backend.conf
terraform apply
```

### Cloud Run Deployment

```bash
# Build and push container
./bin/deploy prd

# Or manually:
docker build -t IMAGE_URL .
docker push IMAGE_URL
gcloud run deploy charaxiv-prd --image IMAGE_URL --region asia-northeast1
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

## Icons

Icons are inline SVGs from Font Awesome, defined in `templates/icons.templ`. This avoids external dependencies and only ships icons you actually use.

### Adding a New Icon

1. Find the icon on [Font Awesome GitHub](https://github.com/FortAwesome/Font-Awesome/tree/6.x/svgs/solid)
2. Get the raw SVG:
   ```bash
   curl -s "https://raw.githubusercontent.com/FortAwesome/Font-Awesome/6.x/svgs/solid/ICON-NAME.svg"
   ```
3. Add to `templates/icons.templ`:
   ```go
   templ IconMyIcon() {
       <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 XXX 512" fill="currentColor" class="icon">
           <path d="..."></path>
       </svg>
   }
   ```
4. Use in templates:
   ```go
   @IconMyIcon()
   ```

### Icon Styling

- Icons use `currentColor`, so they inherit text color from parent
- Default size is `1em` (relative to font-size)
- Override size with custom class or inline style:
  ```css
  .my-large-icon .icon {
      width: 24px;
      height: 24px;
  }
  ```

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
