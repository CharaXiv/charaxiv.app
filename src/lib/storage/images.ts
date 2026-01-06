/**
 * Image storage for R2.
 * Images are stored as binary objects in R2 with metadata.
 */

import type { CharacterImage } from '../types'

export interface ImageStore {
  /**
   * Upload an image.
   * Returns the R2 key for the uploaded image.
   */
  upload(characterId: string, imageId: string, data: ArrayBuffer, contentType: string, filename: string): Promise<string>

  /**
   * Get an image URL or data.
   * Returns null if not found.
   */
  get(characterId: string, imageId: string): Promise<{ data: ArrayBuffer; contentType: string } | null>

  /**
   * Delete an image.
   */
  delete(characterId: string, imageId: string): Promise<void>

  /**
   * Generate a public URL for an image (if bucket is public).
   */
  publicUrl(characterId: string, imageId: string): string
}

/**
 * Create an image store backed by R2.
 */
export function createImageStore(bucket: R2Bucket): ImageStore {
  return {
    async upload(characterId: string, imageId: string, data: ArrayBuffer, contentType: string, filename: string) {
      const key = `images/${characterId}/${imageId}`
      await bucket.put(key, data, {
        httpMetadata: { contentType },
        customMetadata: { filename },
      })
      return key
    },

    async get(characterId: string, imageId: string) {
      const key = `images/${characterId}/${imageId}`
      const object = await bucket.get(key)
      if (!object) return null

      const data = await object.arrayBuffer()
      const contentType = object.httpMetadata?.contentType || 'application/octet-stream'
      return { data, contentType }
    },

    async delete(characterId: string, imageId: string) {
      const key = `images/${characterId}/${imageId}`
      await bucket.delete(key)
    },

    publicUrl(characterId: string, imageId: string) {
      // Note: This assumes the bucket is configured for public access
      // In production, you'd use signed URLs or a custom domain
      return `/api/image/${characterId}/${imageId}`
    },
  }
}
