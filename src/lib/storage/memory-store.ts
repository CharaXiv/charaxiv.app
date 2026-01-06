/**
 * In-memory implementation of AsyncStore for local development.
 * Wraps the existing synchronous store functions.
 */

import type { AsyncStore } from './async-store'
import * as syncStore from '../store'

/**
 * Create an in-memory async store that wraps the synchronous store.
 */
export function createMemoryStore(): AsyncStore {
  return {
    // Status
    async getStatus(charId) {
      return syncStore.getStatus(charId)
    },
    async setVariableBase(charId, key, value) {
      return syncStore.setVariableBase(charId, key, value)
    },
    async setParameter(charId, key, value) {
      return syncStore.setParameter(charId, key, value)
    },
    async adjustParameter(charId, key, delta) {
      return syncStore.adjustParameter(charId, key, delta)
    },
    async setDamageBonus(charId, value) {
      return syncStore.setDamageBonus(charId, value)
    },
    async randomizeStats(charId) {
      return syncStore.randomizeStats(charId)
    },

    // Skills
    async getSkills(charId) {
      return syncStore.getSkills(charId)
    },
    async getSkill(charId, key) {
      return syncStore.getSkill(charId, key)
    },
    async updateSkill(charId, key, updates) {
      return syncStore.updateSkill(charId, key, updates)
    },
    async setExtraJob(charId, value) {
      return syncStore.setExtraJob(charId, value)
    },
    async setExtraHobby(charId, value) {
      return syncStore.setExtraHobby(charId, value)
    },
    async adjustExtraJob(charId, delta) {
      return syncStore.adjustExtraJob(charId, delta)
    },
    async adjustExtraHobby(charId, delta) {
      return syncStore.adjustExtraHobby(charId, delta)
    },

    // Multi-genre skills
    async addGenre(charId, skillKey) {
      return syncStore.addGenre(charId, skillKey)
    },
    async deleteGenre(charId, skillKey, genreIndex) {
      return syncStore.deleteGenre(charId, skillKey, genreIndex)
    },
    async updateGenreGrow(charId, skillKey, genreIndex, grow) {
      return syncStore.updateGenreGrow(charId, skillKey, genreIndex, grow)
    },
    async updateGenreLabel(charId, skillKey, genreIndex, label) {
      return syncStore.updateGenreLabel(charId, skillKey, genreIndex, label)
    },
    async updateGenreField(charId, skillKey, genreIndex, field, delta) {
      return syncStore.updateGenreField(charId, skillKey, genreIndex, field, delta)
    },

    // Custom skills
    async addCustomSkill(charId) {
      return syncStore.addCustomSkill(charId)
    },
    async deleteCustomSkill(charId, index) {
      return syncStore.deleteCustomSkill(charId, index)
    },
    async updateCustomSkillGrow(charId, index, grow) {
      return syncStore.updateCustomSkillGrow(charId, index, grow)
    },
    async updateCustomSkillName(charId, index, name) {
      return syncStore.updateCustomSkillName(charId, index, name)
    },
    async updateCustomSkillField(charId, index, field, delta) {
      return syncStore.updateCustomSkillField(charId, index, field, delta)
    },
    async getCustomSkills(charId) {
      return syncStore.getCustomSkills(charId)
    },

    // Memos
    async getMemos(charId) {
      return syncStore.getMemos(charId)
    },
    async setMemo(charId, id, value) {
      return syncStore.setMemo(charId, id, value)
    },

    // Profile
    async getProfile(charId) {
      return syncStore.getProfile(charId)
    },
    async setProfileName(charId, name) {
      return syncStore.setProfileName(charId, name)
    },
    async setProfileRuby(charId, ruby) {
      return syncStore.setProfileRuby(charId, ruby)
    },

    // Image gallery
    async getImageGallery(charId) {
      return syncStore.getImageGallery(charId)
    },
    async getCurrentImage(charId) {
      return syncStore.getCurrentImage(charId)
    },
    async addImage(charId, filename, data, contentType) {
      return syncStore.addImage(charId, filename, data, contentType)
    },
    async deleteCurrentImage(charId) {
      return syncStore.deleteCurrentImage(charId)
    },
    async pinCurrentImage(charId) {
      return syncStore.pinCurrentImage(charId)
    },
    async navigateImage(charId, direction) {
      return syncStore.navigateImage(charId, direction)
    },
    async setCurrentImageIndex(charId, index) {
      return syncStore.setCurrentImageIndex(charId, index)
    },
  }
}
