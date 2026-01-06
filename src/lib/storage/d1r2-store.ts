/**
 * D1/R2 implementation of AsyncStore for Cloudflare Workers.
 * Uses the CoalesceStore + CharacterService + StoreAdapter pattern.
 */

import type { AsyncStore } from './async-store'
import { createCoalesceStore } from './coalesce'
import { CharacterService } from './character'
import { StoreAdapter } from './store-adapter'

/**
 * Create a D1/R2 backed async store.
 */
export function createD1R2Store(db: D1Database, bucket: R2Bucket): AsyncStore {
  const coalesceStore = createCoalesceStore(db, bucket)
  const service = new CharacterService(coalesceStore)
  const adapter = new StoreAdapter(service, coalesceStore)

  return {
    // Status
    async getStatus(charId) {
      return adapter.getStatus(charId)
    },
    async setVariableBase(charId, key, value) {
      return adapter.setVariableBase(charId, key, value)
    },
    async setParameter(charId, key, value) {
      return adapter.setParameter(charId, key, value)
    },
    async adjustParameter(charId, key, delta) {
      return adapter.adjustParameter(charId, key, delta)
    },
    async setDamageBonus(charId, value) {
      return adapter.setDamageBonus(charId, value)
    },
    async randomizeStats(charId) {
      return adapter.randomizeStats(charId)
    },

    // Skills
    async getSkills(charId) {
      return adapter.getSkills(charId)
    },
    async getSkill(charId, key) {
      const skills = await adapter.getSkills(charId)
      for (const cat of skills.categories) {
        const skill = cat.skills.find(s => s.key === key)
        if (skill) return skill
      }
      return null
    },
    async updateSkill(charId, key, updates) {
      return adapter.updateSkill(charId, key, updates)
    },
    async setExtraJob(charId, value) {
      return adapter.setExtraJob(charId, value)
    },
    async setExtraHobby(charId, value) {
      return adapter.setExtraHobby(charId, value)
    },
    async adjustExtraJob(charId, delta) {
      return adapter.adjustExtraJob(charId, delta)
    },
    async adjustExtraHobby(charId, delta) {
      return adapter.adjustExtraHobby(charId, delta)
    },

    // Multi-genre skills
    async addGenre(charId, skillKey) {
      return adapter.addGenre(charId, skillKey)
    },
    async deleteGenre(charId, skillKey, genreIndex) {
      return adapter.deleteGenre(charId, skillKey, genreIndex)
    },
    async updateGenreGrow(charId, skillKey, genreIndex, grow) {
      return adapter.updateGenreGrow(charId, skillKey, genreIndex, grow)
    },
    async updateGenreLabel(charId, skillKey, genreIndex, label) {
      return adapter.updateGenreLabel(charId, skillKey, genreIndex, label)
    },
    async updateGenreField(charId, skillKey, genreIndex, field, delta) {
      return adapter.updateGenreField(charId, skillKey, genreIndex, field, delta)
    },

    // Custom skills
    async addCustomSkill(charId) {
      return adapter.addCustomSkill(charId)
    },
    async deleteCustomSkill(charId, index) {
      return adapter.deleteCustomSkill(charId, index)
    },
    async updateCustomSkillGrow(charId, index, grow) {
      return adapter.updateCustomSkillGrow(charId, index, grow)
    },
    async updateCustomSkillName(charId, index, name) {
      return adapter.updateCustomSkillName(charId, index, name)
    },
    async updateCustomSkillField(charId, index, field, delta) {
      return adapter.updateCustomSkillField(charId, index, field, delta)
    },
    async getCustomSkills(charId) {
      return adapter.getCustomSkills(charId)
    },

    // Memos
    async getMemos(charId) {
      return adapter.getMemos(charId)
    },
    async setMemo(charId, id, value) {
      return adapter.setMemo(charId, id, value)
    },

    // Profile
    async getProfile(charId) {
      return adapter.getProfile(charId)
    },
    async setProfileName(charId, name) {
      return adapter.setProfileName(charId, name)
    },
    async setProfileRuby(charId, ruby) {
      return adapter.setProfileRuby(charId, ruby)
    },

    // Image gallery
    async getImageGallery(charId) {
      return adapter.getImageGallery(charId)
    },
    async getCurrentImage(charId) {
      return adapter.getCurrentImage(charId)
    },
    async addImage(charId, filename, data, contentType) {
      return adapter.addImage(charId, filename, data, contentType)
    },
    async deleteCurrentImage(charId) {
      return adapter.deleteCurrentImage(charId)
    },
    async pinCurrentImage(charId) {
      return adapter.pinCurrentImage(charId)
    },
    async navigateImage(charId, direction) {
      return adapter.navigateImage(charId, direction)
    },
    async setCurrentImageIndex(charId, index) {
      return adapter.setCurrentImageIndex(charId, index)
    },
  }
}
