# CONTRACT.md

By working here—whether eagerly, reluctantly, or under duress—you agree to `CONTRACT.md`.

## Prime Directive

If you spot a violation, contradiction, or anything fishy vis-a-vis `CONTRACT.md`, escalate until resolved.

## Purpose

`CONTRACT.md` places hard caps on complexity and scope. Reason: It is easier to add complexity than to remove it.

Silence is golden on matters not discussed herein.

## Core Agreements

- Only humans edit `CONTRACT.md`. AI may advise but must refuse to edit unless explicitly instructed.
- When AI is instructed to edit, it must:
  - Change only **one** contractual detail per edit
  - Request clarification on any ambiguity before proceeding
- PRs changing `CONTRACT.md` must change **only** `CONTRACT.md`.
- To signal agreement: "I [name] agree to `CONTRACT.md` as countersigned by Dr. Zin, ESQ."

---

## Repository Specific Contracts

### Client-Side Complexity

- **No JS frameworks.** Hono JSX + HTMX only. No React, Vue, Svelte, Solid, etc.
- **Minimal inline JS.** Allowed for: optimistic UI updates, state preservation across swaps, global event listeners (e.g., loading spinner). Must be co-located in `public/components.js`.
- **No client-side state management.** Server is authoritative. No stores, signals, or reactive state libraries.

### HTMX Patterns

- **Prefer OOB swaps** when chunk replacement would lose client-side state (e.g., `<details>` open/closed, scroll position, focus).
- **Prefer chunk replacement** when updates are self-contained with no state to preserve.
- **No cascading client-side updates.** If A updates B which should update C, that logic belongs on the server.

### Backend Complexity

- **D1 for write buffer, R2 for JSON storage.** No other databases.
- **In-memory store acceptable** during local development without D1/R2 bindings.
- **No background jobs.** Cloudflare Workers are request-response only.

### Multi-tenancy & Auth

- **Single user assumed.** Auth and multi-tenancy will be specified later.
- **No OAuth/SSO scaffolding** until explicitly required.

### Real-time & Offline

- **No real-time collaboration.** No WebSocket sync, no CRDTs, no operational transforms.
- **No offline support.** This is a server-rendered web app.

### Undo/History

- **Server-side rollback only.** If undo is needed, it's via server state, not client-side history.

### External Services

- **Cloudflare only.** D1 + R2 for storage. No other external services without explicit discussion.
- **No analytics, no error tracking, no feature flags.** For now.

---

## TBD (Future Relaxations)

These may grow in complexity later:

- Auth: OAuth when multi-user is required
- Durable Objects: For real-time collaboration (if ever)
