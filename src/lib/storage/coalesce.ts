/**
 * Write coalescing storage for Cloudflare Workers.
 * 
 * Pattern:
 * - Writes are buffered in D1 (fast, ~1ms ack)
 * - Reads flush pending writes to R2 (character JSON)
 * - This amortizes R2 write latency into page load
 */

import type { CharacterImage } from '../types'

// Stored character data structure (persisted to R2)
export interface CharacterData {
  profile: { name: string; ruby: string }
  status: {
    variables: Array<{ key: string; base: number; perm: number; temp: number }>
    parameters: Array<{ key: string; value: number | null }>
    damageBonus: string | null
  }
  skills: {
    single: Record<string, { job: number; hobby: number; perm: number; temp: number; grow: boolean }>
    multi: Record<string, { genres: Array<{ label: string; job: number; hobby: number; perm: number; temp: number; grow: boolean }> }>
    custom: Array<{ name: string; job: number; hobby: number; perm: number; temp: number; grow: boolean }>
    extra: { job: number; hobby: number }
  }
  memos: Record<string, string>
  images: CharacterImage[]
  currentImageIndex: number
}

export interface CoalesceStore {
  /**
   * Buffer a write operation.
   * Fast (~1ms) - writes to D1 only.
   */
  write(characterId: string, path: string, value: unknown): Promise<void>

  /**
   * Read character data, flushing any pending writes.
   * Slower on first load if there are pending writes.
   */
  read(characterId: string): Promise<CharacterData | null>

  /**
   * Check if a character exists.
   */
  exists(characterId: string): Promise<boolean>

  /**
   * Create a new character with default data.
   */
  create(characterId: string): Promise<CharacterData>

  /**
   * Delete a character.
   */
  delete(characterId: string): Promise<void>
}

/**
 * Create a coalescing store backed by D1 (buffer) and R2 (JSON storage).
 */
export function createCoalesceStore(db: D1Database, bucket: R2Bucket): CoalesceStore {
  return {
    async write(characterId: string, path: string, value: unknown) {
      const valueJson = JSON.stringify(value)
      await db.prepare(
        'INSERT INTO write_buffer (character_id, path, value) VALUES (?, ?, ?)'
      ).bind(characterId, path, valueJson).run()
    },

    async read(characterId: string): Promise<CharacterData | null> {
      // Get pending writes from D1
      const pendingResult = await db.prepare(
        'SELECT path, value FROM write_buffer WHERE character_id = ? ORDER BY id'
      ).bind(characterId).all<{ path: string; value: string }>()

      const pending = pendingResult.results || []

      // Load current data from R2
      const r2Key = `characters/${characterId}.json`
      const object = await bucket.get(r2Key)
      
      let data: CharacterData | null = null
      if (object) {
        const text = await object.text()
        data = JSON.parse(text) as CharacterData
      }

      if (!data) {
        return null
      }

      // Apply pending writes
      if (pending.length > 0) {
        for (const write of pending) {
          const value = JSON.parse(write.value)
          setPath(data as unknown as Record<string, unknown>, write.path, value)
        }

        // Save updated data to R2
        await bucket.put(r2Key, JSON.stringify(data), {
          httpMetadata: { contentType: 'application/json' },
        })

        // Clear buffer
        await db.prepare(
          'DELETE FROM write_buffer WHERE character_id = ?'
        ).bind(characterId).run()
      }

      return data
    },

    async exists(characterId: string): Promise<boolean> {
      const r2Key = `characters/${characterId}.json`
      const object = await bucket.head(r2Key)
      return object !== null
    },

    async create(characterId: string): Promise<CharacterData> {
      const data = createDefaultCharacterData()
      const r2Key = `characters/${characterId}.json`
      await bucket.put(r2Key, JSON.stringify(data), {
        httpMetadata: { contentType: 'application/json' },
      })
      return data
    },

    async delete(characterId: string): Promise<void> {
      const r2Key = `characters/${characterId}.json`
      await bucket.delete(r2Key)
      await db.prepare(
        'DELETE FROM write_buffer WHERE character_id = ?'
      ).bind(characterId).run()
    },
  }
}

/**
 * Set a value at a dot-separated path in a nested object.
 * Handles both object keys and array indices (numeric strings).
 * e.g., setPath(data, "skills.single.回避.job", 5)
 *       setPath(data, "status.variables.0.base", 15)
 */
function setPath(data: Record<string, unknown>, path: string, value: unknown): void {
  const parts = path.split('.')
  if (parts.length === 0) return

  let current: unknown = data
  for (let i = 0; i < parts.length - 1; i++) {
    const key = parts[i]
    const nextKey = parts[i + 1]
    const isNextNumeric = /^\d+$/.test(nextKey)
    
    if (Array.isArray(current)) {
      const idx = parseInt(key, 10)
      if (current[idx] === undefined || current[idx] === null) {
        current[idx] = isNextNumeric ? [] : {}
      }
      current = current[idx]
    } else if (typeof current === 'object' && current !== null) {
      const obj = current as Record<string, unknown>
      if (obj[key] === undefined || obj[key] === null) {
        obj[key] = isNextNumeric ? [] : {}
      }
      current = obj[key]
    } else {
      // Can't traverse - this shouldn't happen with valid paths
      return
    }
  }

  // Set final value
  const lastKey = parts[parts.length - 1]
  if (Array.isArray(current)) {
    const idx = parseInt(lastKey, 10)
    current[idx] = value
  } else if (typeof current === 'object' && current !== null) {
    (current as Record<string, unknown>)[lastKey] = value
  }
}

/**
 * Create default character data for CoC 6e.
 */
/**
 * Create default character data for CoC 6e.
 * Defaults are the expected values for each dice roll:
 *   - 3D6 (STR, CON, POW, DEX, APP): expected = 10.5, use 11
 *   - 2D6+6 (SIZ, INT): expected = 13
 *   - 3D6+3 (EDU): expected = 13.5, use 14
 */
function createDefaultCharacterData(): CharacterData {
  return {
    profile: { name: '', ruby: '' },
    status: {
      variables: [
        { key: 'STR', base: 11, perm: 0, temp: 0 },
        { key: 'CON', base: 11, perm: 0, temp: 0 },
        { key: 'POW', base: 11, perm: 0, temp: 0 },
        { key: 'DEX', base: 11, perm: 0, temp: 0 },
        { key: 'APP', base: 11, perm: 0, temp: 0 },
        { key: 'SIZ', base: 13, perm: 0, temp: 0 },
        { key: 'INT', base: 13, perm: 0, temp: 0 },
        { key: 'EDU', base: 14, perm: 0, temp: 0 },
      ],
      parameters: [
        { key: 'HP', value: null },
        { key: 'MP', value: null },
        { key: 'SAN', value: null },
      ],
      damageBonus: null,
    },
    skills: {
      single: {},
      multi: {},
      custom: [],
      extra: { job: 0, hobby: 0 },
    },
    memos: {},
    images: [],
    currentImageIndex: 0,
  }
}

// Re-export the default creator for use elsewhere
export { createDefaultCharacterData }
