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

- **No JS frameworks.** HTMX + templ only. No React, Vue, Svelte, Solid, etc.
- **Minimal inline JS.** Allowed for: optimistic UI updates, state preservation across swaps, global event listeners (e.g., loading spinner). Must be co-located with the element or in a `OnceHandle` script block.
- **No client-side state management.** Server is authoritative. No stores, signals, or reactive state libraries.

### HTMX Patterns

- **Prefer OOB swaps** when chunk replacement would lose client-side state (e.g., `<details>` open/closed, scroll position, focus).
- **Prefer chunk replacement** when updates are self-contained with no state to preserve.
- **No cascading client-side updates.** If A updates B which should update C, that logic belongs on the server.

### Backend Complexity

- **Single SQLite database.** No Redis, no message brokers, no microservices. For now.
- **In-memory store acceptable** during prototyping. Persistence comes later.
- **No background job queues.** Synchronous request-response only. For now.

### Multi-tenancy & Auth

- **Single user assumed.** Auth and multi-tenancy will be specified later.
- **No OAuth/SSO scaffolding** until explicitly required.

### Real-time & Offline

- **No real-time collaboration.** No WebSocket sync, no CRDTs, no operational transforms.
- **No offline support.** This is a server-rendered web app. Offline is a separate native app concern.

### Undo/History

- **Server-side rollback only.** If undo is needed, it's via server state, not client-side history.

### External Services

- **GCS for blob storage.** Already integrated.
- **No other external services** without explicit discussion. No analytics, no error tracking, no feature flags. For now.

---

## TBD (Future Relaxations)

These may grow in complexity later:

- Database: Migration to PostgreSQL if needed
- Auth: OAuth when multi-user is required
- Background jobs: For image processing, exports
- Real-time: For collaborative editing (if ever)
