# CharaXiv Agent Guide

## Domain Model

### Character vs Sheet

A **Character** is the root entity representing a TRPG character. A character can have multiple **Sheets** (also called "systems") for different game systems or scenarios.

This enables players to:
- Use the same character across multiple game systems (CoC 6e, CoC 7e, Emoklore, etc.)
- Have separate sheets for different scenarios within the same system
- Share profile data (images, name) while keeping system-specific data separate

### Data Hierarchy

```
Character
├── Profile
│   ├── name, ruby (reading)
│   ├── images[] (shared across all sheets)
│   ├── pinned image
│   ├── tags[]
│   └── public/secret memos (character-level)
├── Access Control
│   ├── owner
│   ├── editors[]
│   └── viewers[]
└── Sheets[] (one per game system instance)
    ├── id, key (system type), label
    ├── pinned image (can differ from character)
    ├── public/secret memos (scenario-level)
    └── system-specific data (abilities, skills, etc.)
```

### Key Distinction: Memos

- **Character-level memos**: General character background, personality, etc. Shown on the character page.
- **Sheet-level memos**: Scenario-specific notes. Shown on the sheet page. Prefixed with "シナリオ" in UI.

## URL Structure

| Route | Page | Content |
|-------|------|----------|
| `/c/{characterId}` | Character page | Profile + character memos + list of sheets |
| `/c/{characterId}/{systemKey}/{sheetId}` | Sheet page | Profile + scenario memos + system-specific data |

## Current State

The app currently renders a **sheet page** at `/`. This is a placeholder; proper routing with character/sheet IDs will be added with the database layer.

## Legacy Reference

The `~/LegacyCharaXiv` directory contains the previous SvelteKit/Firebase implementation. Useful for:
- Understanding UI patterns and component structure
- Game system data schemas (CoC 6e, CoC 7e, Emoklore)
- Export formats (CCFOLIA, chat palettes)

## Tech Decisions

See `CONTRACT.md` for hard complexity caps.

- **No JS frameworks**: HTMX + templ for interactivity
- **Co-located styles**: CSS in templ files using `OnceHandle`
- **OOB swaps**: For updating multiple page regions from one response

### OOB Swaps vs Chunk Replacement

Prefer **granular OOB swaps** over replacing a large HTML chunk when:
- Client-side state would be lost (e.g., `<details>` open/closed state, scroll position, focus)
- The alternative requires JavaScript to preserve/restore that state

Prefer **chunk replacement** when:
- Updates are self-contained with no client-side state to preserve
- The OOB target list would be excessively long or unpredictable

**Example**: Skill value changes use OOB to update input + breakdown + total + remaining points. This preserves the `<details>` expand/collapse state without client-side JavaScript. Replacing the entire skills panel would collapse open details on every update.

## Templ Gotchas

### Boolean vs Valued Attributes

Templ's `attr?={ bool }` syntax renders a **boolean attribute** (present or absent, no value). This doesn't work for HTMX attributes that require explicit values.

```go
// WRONG: renders `hx-swap-oob` (no value) - HTMX ignores this
<div hx-swap-oob?={ oob }>

// CORRECT: renders `hx-swap-oob="true"` - HTMX processes this
<div
    if oob {
        hx-swap-oob="true"
    }
>
```

Use the `if` block pattern for HTMX attributes like `hx-swap-oob`, `hx-boost`, etc.

## Development

The dev server runs as a systemd service with auto-restart:

```bash
# Check status
sudo systemctl status charaxiv-dev

# Restart (after config changes)
sudo systemctl restart charaxiv-dev

# View logs
journalctl -u charaxiv-dev -f
```

The service handles hot reload automatically - just edit files and the browser will refresh.

See `README.md` for:
- Build commands
- Infrastructure (Terraform, GCS, Cloud Run)
- Design system tokens (colors, spacing, typography)
- Component patterns and gotchas
