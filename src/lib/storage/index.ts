/**
 * Storage module for Cloudflare Workers.
 * Provides D1 (write buffer) + R2 (JSON/images) backed storage.
 */

export { createCoalesceStore, createDefaultCharacterData } from './coalesce'
export type { CoalesceStore, CharacterData } from './coalesce'

export { createImageStore } from './images'
export type { ImageStore } from './images'

export { CharacterService } from './character'
export { StoreAdapter } from './store-adapter'

// Unified async store interface
export type { AsyncStore } from './async-store'
export { createMemoryStore } from './memory-store'
export { createD1R2Store } from './d1r2-store'
