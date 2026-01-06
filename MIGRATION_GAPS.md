# CharaXiv Hono Migration - Feature Gap Analysis

This document compares the original Go/templ CharaXiv implementation with the Hono/JSX migration to identify missing features and functionality gaps.

## Summary

The Hono implementation now has complete core functionality. **P0, P1, P2, and P3 items are COMPLETE** (except gallery modal view and design token audit, which are low priority).

### Completed Features:
1. ✅ **Skill Editing UI** - Single skills, multi-genre skills, and custom skills all working
2. ✅ **API Routes** - All skill-related endpoints implemented with OOB responses
3. ✅ **OOB Swap Patterns** - Granular updates for all skill operations (preserves `<details>` state)
4. ✅ **EasyMDE Integration** - Markdown editor loads and syncs
5. ✅ **Preview Mode** - OOB fragments returning complete state
6. ✅ **Parameter Adjustment** - HP/MP/SAN ±1/±5 buttons with OOB updates
7. ✅ **Random Stats** - Full CoC 6e dice rules implementation (3d6, 2d6+6, 3d6+3)
8. ✅ **Profile Auto-save** - Name/ruby inputs with HTMX debounce
9. ✅ **Damage Bonus Input** - Text input with auto-save
10. ✅ **Markdown Rendering** - marked.js renders markdown in preview mode
11. ✅ **Secret Memo Toggle** - Blur/reveal toggle for secret memos in preview mode
12. ✅ **Image Gallery** - Upload/delete/pin/navigation with HTMX OOB updates

### Remaining Gaps (P2/P3):
1. ✅ ~~Parameter adjustment (HP/MP/SAN)~~ - IMPLEMENTED
2. ✅ ~~Profile input auto-save~~ - IMPLEMENTED
3. ✅ ~~Markdown rendering in read-only mode~~ - IMPLEMENTED (uses marked.js)
4. ✅ ~~Random stats button~~ - IMPLEMENTED (CoC 6e dice rules)
5. ✅ ~~Image gallery functionality~~ - IMPLEMENTED (upload/delete/pin/navigation)
6. ✅ ~~Damage bonus text input~~ - IMPLEMENTED
7. ✅ ~~D1 integration for write buffer~~ - IMPLEMENTED (schema + coalesce module)
8. ✅ ~~R2 integration for character JSON~~ - IMPLEMENTED (storage module)
9. ✅ ~~Wrangler deployment config~~ - IMPLEMENTED (migrations, bindings)
10. ⚠️ Switch routes to use D1/R2 storage - NOT YET (using in-memory store)
11. ⚠️ Gallery modal view - LOW PRIORITY

---

## 1. Skills Panel - Major Gaps

### 1.1 Multi-Genre Skills (運転, 芸術, ほかの言語, etc.) ✅ IMPLEMENTED

**Original:** Full support for multi-genre skills with:
- Add/delete genre buttons
- Editable genre labels (e.g., "自動車" for 運転)
- Per-genre job/hobby/perm/temp inputs
- Per-genre grow checkbox
- Genre deletion with confirmation

**Hono:** ✅ IMPLEMENTED
- ✅ Genre row rendering with expandable details
- ✅ Add/delete genre functionality with confirmation
- ✅ Editable genre labels (inline text input)
- ✅ All API routes for genres

**Implemented Components:**
- ✅ `SkillMultiRow` - Multi-genre skill container
- ✅ `GenreRow` - Individual genre row with expand/collapse
- ✅ `GenreDetail` - Expanded detail with inputs
- ✅ `GenreGrowButton` - Grow checkbox for genres
- ✅ `GenrePointsBreakdown` - Points display for genres
- ✅ `GenreTotal` - Total display for genres
- ✅ `GenreName` - Editable genre label

**Implemented API Routes:**
```
✅ POST /api/skill/{key}/genre/add
✅ POST /api/skill/{key}/genre/{index}/delete  
✅ POST /api/skill/{key}/genre/{index}/grow
✅ POST /api/skill/{key}/genre/{index}/label
✅ POST /api/skill/{key}/genre/{index}/{field}/adjust
```

**OOB Fragments:**
- ✅ `GenreUpdateFragments` - Updates genre detail, breakdown, total, name, and points
- ✅ `GenreGrowUpdateFragments` - Updates grow button and name styling
- ✅ `SkillMultiPanelFragment` - Updates entire multi-skill section (for add/delete)

### 1.2 Custom Skills (独自技能) ✅ IMPLEMENTED

**Original:** Full support for user-defined skills:
- Add custom skill button
- Editable skill name
- Delete with confirmation
- Grow checkbox
- Job/hobby/perm/temp inputs

**Hono:** ✅ IMPLEMENTED
- ✅ Custom skills section with add button
- ✅ All API routes for custom skills
- ✅ Store functions for custom skill management

**Implemented Components:**
- ✅ `CustomSkillsSection` - Section header with add button
- ✅ `CustomSkillRow` - Expandable row with delete button
- ✅ `CustomSkillDetail` - Detail inputs  
- ✅ `CustomSkillGrowButton` - Grow checkbox
- ✅ `CustomSkillPointsBreakdown` - Points breakdown display
- ✅ `CustomSkillTotal` - Total display
- ✅ `CustomSkillName` - Editable skill name

**Implemented API Routes:**
```
✅ POST /api/skill/custom/add
✅ POST /api/skill/custom/{index}/delete
✅ POST /api/skill/custom/{index}/grow
✅ POST /api/skill/custom/{index}/name
✅ POST /api/skill/custom/{index}/{field}/adjust
```

### 1.3 Skill Row Interactivity

**Original:** Full interactive skill rows:
- Expandable `<details>` with preserved open/close state
- Grow button toggles independently via OOB swap
- Name styling updates (active/inactive) via OOB
- Points breakdown updates via OOB
- Total updates via OOB

**Hono:** ✅ IMPLEMENTED
- ✅ Basic expandable rows
- ✅ Grow button updates via OOB (SkillGrowUpdateFragments)
- ✅ Skill adjustment uses OOB fragments (preserves `<details>` state)
- ✅ OOB fragments for granular updates

**Implemented OOB Fragment Components:**
- ✅ `SkillGrowUpdateFragments` - Returns grow button + skill name styling
- ✅ `SkillUpdateFragments` - Returns input + breakdown + total + points display

### 1.4 Extra Skill Points

**Original:**
- Extra job/hobby point inputs in skills panel
- Updates trigger OOB refresh of remaining points display

**Hono:** ✅ IMPLEMENTED
- ✅ Inputs rendered
- ✅ API routes for set/adjust implemented
- ✅ Store functions (setExtraJob, setExtraHobby, adjustExtraJob, adjustExtraHobby)
- ✅ OOB response with ExtraPointsUpdateFragments

---

## 2. Status Panel - ✅ COMPLETE

### 2.1 Parameter Adjustment ✅ IMPLEMENTED

**Original:**
- HP/MP/SAN can be adjusted with ±1/±5 buttons
- Store tracks overrides vs defaults
- Updates return OOB status panel

**Hono:** ✅ IMPLEMENTED
- ✅ Inputs rendered
- ✅ Adjustment routes handle `param-*` keys with proper persistence
- ✅ `setParameter` and `adjustParameter` functions in store

### 2.2 Damage Bonus Input ✅ IMPLEMENTED

**Original:**
- Text input for DB with HTMX posting
- Server-side validation

**Hono:** ✅ IMPLEMENTED
- ✅ Input rendered
- ✅ API route `/api/param/db/set` with HTMX auto-save
- ✅ Store tracks override vs computed value

### 2.3 Random Button ✅ IMPLEMENTED

**Original:**
- Random button triggers dice roll for all stats
- Returns OOB status panel + skill panel (for DEX/EDU changes)

**Hono:** ✅ IMPLEMENTED
- ✅ Button rendered with hx-post
- ✅ API route `/api/status/random`
- ✅ Randomization logic using CoC 6e dice rules:
  - STR, CON, POW, DEX, APP: 3d6
  - SIZ, INT: 2d6+6
  - EDU: 3d6+3
- ✅ Returns OOB status panel + skills panel + points display

---

## 3. Markdown Editor - Gaps

### 3.1 EasyMDE Integration

**Original:**
- Full EasyMDE rich text editor
- Toolbar: heading, quote, lists, bold, italic, strikethrough, link
- Secret memos have password button in toolbar
- Auto-save with 3s debounce
- Content synced to hidden textarea

**Hono:** ✅ IMPLEMENTED
- ✅ EasyMDE script loaded from CDN in Layout
- ✅ EasyMDE initialization in components.js
- ✅ Full toolbar customization (heading, quote, lists, bold, italic, strikethrough, link)
- ✅ Secret memos have password button in toolbar
- ✅ Auto-save with 3s debounce via HTMX
- ✅ Content synced to hidden textarea

### 3.2 Markdown Rendering (Read-only Mode) ✅ IMPLEMENTED

**Original:**
- Uses `marked.js` to render markdown
- Secret memos have blur + toggle button
- HTMX afterSwap triggers re-render

**Hono:** ✅ IMPLEMENTED
- ✅ marked.js loaded from CDN in Layout
- ✅ `renderMarkdown()` function in components.js
- ✅ HTMX afterSwap triggers re-render
- ✅ Secret memos have blur + eye icon toggle button
- ✅ Secret toggle JS implemented (blur/reveal toggle working)

---

## 4. Preview Mode - ✅ IMPLEMENTED

**Original:**
- Toggle returns `PreviewModeFragments` with:
  - `MemoGroup` (swapped to target)
  - `ScenarioMemoGroup` (OOB swap)
  - `HeaderActions` (OOB swap to update toggle button state)

**Hono:** ✅ IMPLEMENTED
- ✅ Toggle button renders
- ✅ Returns complete OOB fragments:
  - ✅ MemoGroup (swapped to target)
  - ✅ ScenarioMemoGroup (OOB swap)
  - ✅ SheetHeaderActionsOOB (toggle button updates visually)

---

## 5. Image Gallery - ✅ IMPLEMENTED

**Original:**
- Image display with placeholder
- Navigation (prev/next)
- Upload button with file input
- Delete button with confirmation
- Download link
- Pin button
- Gallery modal

**Hono:**
- ✅ Image display with placeholder
- ✅ Navigation (prev/next) with circular wrapping
- ✅ Upload via file input with HTMX multipart
- ✅ Delete current image (with hx-confirm)
- ✅ Download link (data URL download)
- ✅ Pin button toggle (solid-orange when pinned)
- ⚠️ Gallery modal not implemented (low priority)

**API Routes Implemented:**
- `POST /api/image/upload` - Upload image (stored as base64 in-memory)
- `POST /api/image/delete` - Delete current image
- `POST /api/image/pin` - Toggle pin on current image
- `POST /api/image/prev` - Navigate to previous image
- `POST /api/image/next` - Navigate to next image

---

## 6. Profile Section - ✅ COMPLETE

**Original:**
- Name input with inline editing
- Ruby (furigana) input
- HTMX auto-save

**Hono:** ✅ IMPLEMENTED
- ✅ Inputs render with proper styling
- ✅ HTMX posting with 1.5s debounce
- ✅ API routes for profile updates:
  - `POST /api/profile/name/set`
  - `POST /api/profile/ruby/set`
- ✅ Store functions for profile persistence

---

## 7. Layout/CSS - Differences

**Original (CSS Variables):**
```css
--space-1: 0.25rem
--space-2: 0.5rem
--radius-md: 8px
--blue-500: #3b82f6
/* ... full design system */
```

**Hono (Tailwind):**
- Uses Tailwind utility classes
- Lost some precise spacing/color consistency
- Some component styling differs slightly

**Recommendation:** Audit Tailwind classes against original CSS vars for consistency.

---

## 8. JavaScript Dependencies - Gaps

### Original `components.js` functionality:

```javascript
// Number input adjustment (optimistic UI)
function adjustNumberInput(btn, delta) { ... }

// EasyMDE initialization
function initMarkdownEditors() { ... }

// Markdown rendering
function renderMarkdown() { ... }
```

**Hono:**
- ⚠️ `adjustNumberInput` exists but needs verification
- ❌ No EasyMDE init
- ❌ No markdown rendering

---

## 9. Store Implementation - Gaps

### Missing Store Functions:

```typescript
// Parameters
updateParameter(charId: string, key: string, delta: number): void

// Custom skills
addCustomSkill(charId: string): CustomSkill
getCustomSkill(charId: string, index: number): CustomSkill | null
updateCustomSkill(charId: string, index: number, skill: CustomSkill): void
deleteCustomSkill(charId: string, index: number): boolean

// Extra points
setSkillExtra(charId: string, job: number, hobby: number): void

// Genres
addGenre(charId: string, skillKey: string): void
deleteGenre(charId: string, skillKey: string, index: number): void
updateGenreLabel(charId: string, skillKey: string, index: number, label: string): void
updateGenreGrow(charId: string, skillKey: string, index: number, grow: boolean): void
updateGenreField(charId: string, skillKey: string, index: number, field: string, delta: number): void
```

---

## 10. API Route Coverage - ✅ COMPLETE

| Route | Original | Hono | Status |
|-------|----------|------|--------|
| GET /cthulhu6 | ✅ | ✅ | Complete |
| POST /api/preview/on | ✅ | ✅ | Complete with OOB |
| POST /api/preview/off | ✅ | ✅ | Complete with OOB |
| POST /api/status/{key}/set | ✅ | ✅ | Works (includes param-*) |
| POST /api/status/{key}/adjust | ✅ | ✅ | Works (includes param-*) |
| POST /api/status/random | ✅ | ✅ | Complete |
| POST /api/param/db/set | ✅ | ✅ | Complete |
| POST /api/profile/name/set | ✅ | ✅ | Complete |
| POST /api/profile/ruby/set | ✅ | ✅ | Complete |
| POST /api/memo/{id}/set | ✅ | ✅ | Works |
| POST /api/skill/{key}/grow | ✅ | ✅ | OOB response |
| POST /api/skill/{key}/{field}/adjust | ✅ | ✅ | OOB fragments |
| POST /api/skill/{key}/genre/add | ✅ | ✅ | Complete |
| POST /api/skill/{key}/genre/{index}/delete | ✅ | ✅ | Complete |
| POST /api/skill/{key}/genre/{index}/grow | ✅ | ✅ | Complete |
| POST /api/skill/{key}/genre/{index}/label | ✅ | ✅ | Complete |
| POST /api/skill/{key}/genre/{index}/{field}/adjust | ✅ | ✅ | Complete |
| POST /api/skill/custom/add | ✅ | ✅ | Complete |
| POST /api/skill/custom/{index}/delete | ✅ | ✅ | Complete |
| POST /api/skill/custom/{index}/grow | ✅ | ✅ | Complete |
| POST /api/skill/custom/{index}/name | ✅ | ✅ | Complete |
| POST /api/skill/custom/{index}/{field}/adjust | ✅ | ✅ | Complete |

---

## Priority Implementation Order

### P0 - Critical (Basic functionality broken) ✅ COMPLETE
1. ✅ Fix skill adjust to use OOB fragments (preserve `<details>` state)
2. ✅ Implement extra points adjustment
3. ✅ Fix preview mode OOB fragments
4. ✅ `adjustNumberInput` JS already in components.js
5. ✅ Skill grow OOB fragments

### P1 - High (Major features missing) ✅ COMPLETE
1. ✅ Multi-genre skill UI and API routes - IMPLEMENTED
2. ✅ Custom skills UI and API routes - IMPLEMENTED  
3. ✅ EasyMDE integration (working - editor loads and syncs)

### P2 - Medium (Polish) ✅ COMPLETE
1. ✅ Parameter adjustment (HP/MP/SAN) - IMPLEMENTED
2. ✅ Profile input auto-save - IMPLEMENTED
3. ✅ Markdown rendering in read-only mode - IMPLEMENTED
4. ✅ Random stats button - IMPLEMENTED

### P3 - Low (Nice to have)
1. ✅ Image gallery functionality - IMPLEMENTED
2. ✅ Damage bonus text input - IMPLEMENTED
3. ✅ Secret memo toggle (blur/reveal) - IMPLEMENTED
4. ✅ D1 integration for write buffer - IMPLEMENTED
5. ✅ R2 integration for character JSON - IMPLEMENTED
6. ✅ Wrangler deployment config - IMPLEMENTED
7. Full design token audit - Not done (low priority)
8. ⚠️ Gallery modal view - Not implemented (low priority)
9. ⚠️ Switch routes to use D1/R2 storage - Not done (using in-memory store)

---

## D1/R2 Storage Implementation

### Completed:
- `migrations/0001_init.sql` - D1 schema for write buffer
- `src/lib/storage/coalesce.ts` - Write coalescing store (D1 buffer + R2 JSON)
- `src/lib/storage/images.ts` - Image storage in R2
- `src/lib/storage/index.ts` - Module exports
- `wrangler.toml` - D1/R2 bindings and migrations config

### Pattern:
```
Write path (~1ms):
  POST {path: "skills.single.回避.job", value: 5}
    → INSERT INTO write_buffer (D1)
    → 200 OK

Read path (page load):
  GET /cthulhu6
    → Load JSON from R2
    → Apply pending writes from D1 buffer
    → Save updated JSON to R2 (flush)
    → Clear D1 buffer
    → Return response
```

### Remaining:
- Switch routes from in-memory `store.ts` to D1/R2 `storage/coalesce.ts`
- This is primarily a refactoring task, not new functionality

---

## Files Created/Modified

### New Files:
- `migrations/0001_init.sql` - D1 migration
- `src/lib/storage/coalesce.ts` - Write coalescing store
- `src/lib/storage/images.ts` - Image storage
- `src/lib/storage/index.ts` - Module exports
- `src/lib/storage/schema.sql` - Schema reference

### Modified Files:
- `wrangler.toml` - Added migrations_dir config
