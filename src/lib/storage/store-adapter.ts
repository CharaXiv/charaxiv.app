/**
 * Store adapter that provides the same interface as store.ts
 * but uses D1/R2 storage via the coalesce pattern.
 */

import type { CoalesceStore, CharacterData } from './coalesce'
import type { CharacterService } from './character'
import type { 
  StatusState, SkillsState, StatusVariable, StatusParameter, 
  SkillCategory, Skill, CustomSkill, CharacterImage, ImageGalleryState
} from '../types'

/**
 * Adapter that wraps CharacterService to provide store-like interface.
 * This is a "fat read, thin write" pattern:
 * - Reads load full data from R2, hydrate with defaults
 * - Writes buffer to D1, flush on next read
 */
export class StoreAdapter {
  private cache: Map<string, CharacterData> = new Map()

  constructor(
    private service: CharacterService,
    private store: CoalesceStore
  ) {}

  /**
   * Load character data (with caching for single request).
   */
  async load(charId: string): Promise<CharacterData> {
    if (this.cache.has(charId)) {
      return this.cache.get(charId)!
    }
    const data = await this.service.getOrCreate(charId)
    this.cache.set(charId, data)
    return data
  }

  /**
   * Invalidate cache (call after writes that need immediate read-back).
   */
  invalidate(charId: string): void {
    this.cache.delete(charId)
  }

  // ========== Status Functions ==========

  async getStatus(charId: string): Promise<StatusState> {
    const data = await this.load(charId)
    return this.service.getStatus(data, data.skills.extra)
  }

  async setVariableBase(charId: string, key: string, value: number): Promise<StatusVariable | null> {
    const data = await this.load(charId)
    const varIdx = data.status.variables.findIndex(v => v.key === key)
    if (varIdx === -1) return null

    await this.store.write(charId, `status.variables.${varIdx}.base`, value)
    data.status.variables[varIdx].base = value
    
    const status = this.service.getStatus(data, data.skills.extra)
    return status.variables.find(v => v.key === key) || null
  }

  async setParameter(charId: string, key: string, value: number): Promise<boolean> {
    const data = await this.load(charId)
    const paramIdx = data.status.parameters.findIndex(p => p.key === key)
    if (paramIdx === -1) return false

    await this.store.write(charId, `status.parameters.${paramIdx}.value`, value)
    data.status.parameters[paramIdx].value = value
    return true
  }

  async adjustParameter(charId: string, key: string, delta: number): Promise<boolean> {
    const data = await this.load(charId)
    const param = data.status.parameters.find(p => p.key === key)
    if (!param) return false

    const status = this.service.getStatus(data, data.skills.extra)
    const statusParam = status.parameters.find(p => p.key === key)
    if (!statusParam) return false

    const current = param.value ?? statusParam.defaultValue
    const newValue = Math.max(0, current + delta)
    return this.setParameter(charId, key, newValue)
  }

  async setDamageBonus(charId: string, value: string): Promise<boolean> {
    await this.store.write(charId, 'status.damageBonus', value)
    const data = await this.load(charId)
    data.status.damageBonus = value
    return true
  }

  async randomizeStats(charId: string): Promise<boolean> {
    const data = await this.load(charId)
    
    const roll3d6 = () => Math.floor(Math.random() * 6) + 1 + Math.floor(Math.random() * 6) + 1 + Math.floor(Math.random() * 6) + 1
    const roll2d6plus6 = () => Math.floor(Math.random() * 6) + 1 + Math.floor(Math.random() * 6) + 1 + 6
    const roll3d6plus3 = () => roll3d6() + 3

    for (let i = 0; i < data.status.variables.length; i++) {
      const v = data.status.variables[i]
      let newBase: number
      switch (v.key) {
        case 'STR': case 'CON': case 'POW': case 'DEX': case 'APP':
          newBase = roll3d6()
          break
        case 'SIZ': case 'INT':
          newBase = roll2d6plus6()
          break
        case 'EDU':
          newBase = roll3d6plus3()
          break
        default:
          continue
      }
      await this.store.write(charId, `status.variables.${i}.base`, newBase)
      v.base = newBase
    }
    return true
  }

  // ========== Skills Functions ==========

  async getSkills(charId: string): Promise<SkillsState> {
    const data = await this.load(charId)
    const status = this.service.getStatus(data, data.skills.extra)
    return this.service.getSkills(data, status.variables)
  }

  async getSkill(charId: string, key: string): Promise<Skill | null> {
    const skills = await this.getSkills(charId)
    for (const cat of skills.categories) {
      const skill = cat.skills.find(s => s.key === key)
      if (skill) return skill
    }
    return null
  }

  async updateSkill(charId: string, key: string, updates: Partial<{ job: number; hobby: number; perm: number; temp: number; grow: boolean }>): Promise<Skill | null> {
    const data = await this.load(charId)
    const current = data.skills.single[key] || { job: 0, hobby: 0, perm: 0, temp: 0, grow: false }
    const updated = { ...current, ...updates }
    
    await this.store.write(charId, `skills.single.${key}`, updated)
    data.skills.single[key] = updated
    
    return this.getSkill(charId, key)
  }

  async setExtraJob(charId: string, value: number): Promise<boolean> {
    const data = await this.load(charId)
    const newValue = Math.max(0, value)
    await this.store.write(charId, 'skills.extra.job', newValue)
    data.skills.extra.job = newValue
    return true
  }

  async setExtraHobby(charId: string, value: number): Promise<boolean> {
    const data = await this.load(charId)
    const newValue = Math.max(0, value)
    await this.store.write(charId, 'skills.extra.hobby', newValue)
    data.skills.extra.hobby = newValue
    return true
  }

  async adjustExtraJob(charId: string, delta: number): Promise<boolean> {
    const data = await this.load(charId)
    return this.setExtraJob(charId, data.skills.extra.job + delta)
  }

  async adjustExtraHobby(charId: string, delta: number): Promise<boolean> {
    const data = await this.load(charId)
    return this.setExtraHobby(charId, data.skills.extra.hobby + delta)
  }

  // ========== Multi-Genre Skills ==========

  async addGenre(charId: string, skillKey: string): Promise<{ label: string; job: number; hobby: number; perm: number; temp: number; grow: boolean } | null> {
    const data = await this.load(charId)
    const multi = data.skills.multi[skillKey]
    if (!multi) {
      // Initialize multi skill
      data.skills.multi[skillKey] = { genres: [] }
    }
    
    const newGenre = { label: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }
    const genres = data.skills.multi[skillKey].genres
    genres.push(newGenre)
    
    await this.store.write(charId, `skills.multi.${skillKey}.genres`, genres)
    return newGenre
  }

  async deleteGenre(charId: string, skillKey: string, genreIndex: number): Promise<boolean> {
    const data = await this.load(charId)
    const multi = data.skills.multi[skillKey]
    if (!multi || multi.genres.length <= 1) return false
    if (genreIndex < 0 || genreIndex >= multi.genres.length) return false

    multi.genres.splice(genreIndex, 1)
    await this.store.write(charId, `skills.multi.${skillKey}.genres`, multi.genres)
    return true
  }

  async updateGenreGrow(charId: string, skillKey: string, genreIndex: number, grow: boolean): Promise<boolean> {
    const data = await this.load(charId)
    const multi = data.skills.multi[skillKey]
    if (!multi || genreIndex < 0 || genreIndex >= multi.genres.length) return false

    multi.genres[genreIndex].grow = grow
    await this.store.write(charId, `skills.multi.${skillKey}.genres.${genreIndex}.grow`, grow)
    return true
  }

  async updateGenreLabel(charId: string, skillKey: string, genreIndex: number, label: string): Promise<boolean> {
    const data = await this.load(charId)
    const multi = data.skills.multi[skillKey]
    if (!multi || genreIndex < 0 || genreIndex >= multi.genres.length) return false

    multi.genres[genreIndex].label = label
    await this.store.write(charId, `skills.multi.${skillKey}.genres.${genreIndex}.label`, label)
    return true
  }

  async updateGenreField(charId: string, skillKey: string, genreIndex: number, field: 'job' | 'hobby' | 'perm' | 'temp', delta: number): Promise<boolean> {
    const data = await this.load(charId)
    const multi = data.skills.multi[skillKey]
    if (!multi || genreIndex < 0 || genreIndex >= multi.genres.length) return false

    const genre = multi.genres[genreIndex]
    const newValue = Math.max(0, genre[field] + delta)
    genre[field] = newValue
    await this.store.write(charId, `skills.multi.${skillKey}.genres.${genreIndex}.${field}`, newValue)
    return true
  }

  // ========== Custom Skills ==========

  async addCustomSkill(charId: string): Promise<CustomSkill | null> {
    const data = await this.load(charId)
    const newSkill: CustomSkill = { name: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }
    data.skills.custom.push(newSkill)
    await this.store.write(charId, 'skills.custom', data.skills.custom)
    return newSkill
  }

  async deleteCustomSkill(charId: string, index: number): Promise<boolean> {
    const data = await this.load(charId)
    if (index < 0 || index >= data.skills.custom.length) return false
    data.skills.custom.splice(index, 1)
    await this.store.write(charId, 'skills.custom', data.skills.custom)
    return true
  }

  async updateCustomSkillGrow(charId: string, index: number, grow: boolean): Promise<boolean> {
    const data = await this.load(charId)
    if (index < 0 || index >= data.skills.custom.length) return false
    data.skills.custom[index].grow = grow
    await this.store.write(charId, `skills.custom.${index}.grow`, grow)
    return true
  }

  async updateCustomSkillName(charId: string, index: number, name: string): Promise<boolean> {
    const data = await this.load(charId)
    if (index < 0 || index >= data.skills.custom.length) return false
    data.skills.custom[index].name = name
    await this.store.write(charId, `skills.custom.${index}.name`, name)
    return true
  }

  async updateCustomSkillField(charId: string, index: number, field: 'job' | 'hobby' | 'perm' | 'temp', delta: number): Promise<boolean> {
    const data = await this.load(charId)
    if (index < 0 || index >= data.skills.custom.length) return false
    const skill = data.skills.custom[index]
    const newValue = Math.max(0, skill[field] + delta)
    skill[field] = newValue
    await this.store.write(charId, `skills.custom.${index}.${field}`, newValue)
    return true
  }

  async getCustomSkills(charId: string): Promise<CustomSkill[]> {
    const data = await this.load(charId)
    return data.skills.custom
  }

  // ========== Memos ==========

  async getMemos(charId: string): Promise<Record<string, string>> {
    const data = await this.load(charId)
    return data.memos
  }

  async setMemo(charId: string, id: string, value: string): Promise<boolean> {
    const data = await this.load(charId)
    data.memos[id] = value
    await this.store.write(charId, `memos.${id}`, value)
    return true
  }

  // ========== Profile ==========

  async getProfile(charId: string): Promise<{ name: string; ruby: string }> {
    const data = await this.load(charId)
    return data.profile
  }

  async setProfileName(charId: string, name: string): Promise<boolean> {
    const data = await this.load(charId)
    data.profile.name = name
    await this.store.write(charId, 'profile.name', name)
    return true
  }

  async setProfileRuby(charId: string, ruby: string): Promise<boolean> {
    const data = await this.load(charId)
    data.profile.ruby = ruby
    await this.store.write(charId, 'profile.ruby', ruby)
    return true
  }

  // ========== Image Gallery ==========

  async getImageGallery(charId: string): Promise<ImageGalleryState> {
    const data = await this.load(charId)
    return this.service.getImageGallery(data)
  }

  async getCurrentImage(charId: string): Promise<CharacterImage | null> {
    const data = await this.load(charId)
    if (data.images.length === 0) return null
    const index = Math.min(data.currentImageIndex, data.images.length - 1)
    return data.images[index]
  }

  async addImage(charId: string, filename: string, dataUrl: string, contentType: string): Promise<CharacterImage> {
    const data = await this.load(charId)
    const image: CharacterImage = {
      id: 'img_' + Date.now().toString(36) + Math.random().toString(36).substr(2, 9),
      filename,
      data: dataUrl,
      contentType,
      pinned: false,
    }
    data.images.push(image)
    data.currentImageIndex = data.images.length - 1
    
    await this.store.write(charId, 'images', data.images)
    await this.store.write(charId, 'currentImageIndex', data.currentImageIndex)
    return image
  }

  async deleteCurrentImage(charId: string): Promise<boolean> {
    const data = await this.load(charId)
    if (data.images.length === 0) return false
    
    data.images.splice(data.currentImageIndex, 1)
    if (data.images.length === 0) {
      data.currentImageIndex = 0
    } else if (data.currentImageIndex >= data.images.length) {
      data.currentImageIndex = data.images.length - 1
    }
    
    await this.store.write(charId, 'images', data.images)
    await this.store.write(charId, 'currentImageIndex', data.currentImageIndex)
    return true
  }

  async pinCurrentImage(charId: string): Promise<boolean> {
    const data = await this.load(charId)
    if (data.images.length === 0) return false
    
    const currentImage = data.images[data.currentImageIndex]
    // Toggle pin, unpin all others
    for (const img of data.images) {
      img.pinned = img.id === currentImage.id ? !img.pinned : false
    }
    
    await this.store.write(charId, 'images', data.images)
    return true
  }

  async navigateImage(charId: string, direction: 'prev' | 'next'): Promise<boolean> {
    const data = await this.load(charId)
    if (data.images.length <= 1) return false
    
    if (direction === 'prev') {
      data.currentImageIndex = (data.currentImageIndex - 1 + data.images.length) % data.images.length
    } else {
      data.currentImageIndex = (data.currentImageIndex + 1) % data.images.length
    }
    
    await this.store.write(charId, 'currentImageIndex', data.currentImageIndex)
    return true
  }

  async setCurrentImageIndex(charId: string, index: number): Promise<boolean> {
    const data = await this.load(charId)
    if (index < 0 || index >= data.images.length) return false
    data.currentImageIndex = index
    await this.store.write(charId, 'currentImageIndex', index)
    return true
  }
}
