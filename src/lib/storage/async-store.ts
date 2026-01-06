/**
 * Async store interface for character data.
 * Both in-memory (dev) and D1/R2 (Workers) implementations provide this interface.
 */

import type { 
  StatusState, SkillsState, StatusVariable, Skill, CustomSkill, 
  CharacterImage, ImageGalleryState, SingleSkillData, SkillGenre
} from '../types'

export interface AsyncStore {
  // Status
  getStatus(charId: string): Promise<StatusState>
  setVariableBase(charId: string, key: string, value: number): Promise<StatusVariable | null>
  setParameter(charId: string, key: string, value: number): Promise<boolean>
  adjustParameter(charId: string, key: string, delta: number): Promise<boolean>
  setDamageBonus(charId: string, value: string): Promise<boolean>
  randomizeStats(charId: string): Promise<boolean>

  // Skills
  getSkills(charId: string): Promise<SkillsState>
  getSkill(charId: string, key: string): Promise<Skill | null>
  updateSkill(charId: string, key: string, updates: Partial<SingleSkillData>): Promise<Skill | null>
  setExtraJob(charId: string, value: number): Promise<boolean>
  setExtraHobby(charId: string, value: number): Promise<boolean>
  adjustExtraJob(charId: string, delta: number): Promise<boolean>
  adjustExtraHobby(charId: string, delta: number): Promise<boolean>

  // Multi-genre skills
  addGenre(charId: string, skillKey: string): Promise<SkillGenre | null>
  deleteGenre(charId: string, skillKey: string, genreIndex: number): Promise<boolean>
  updateGenreGrow(charId: string, skillKey: string, genreIndex: number, grow: boolean): Promise<boolean>
  updateGenreLabel(charId: string, skillKey: string, genreIndex: number, label: string): Promise<boolean>
  updateGenreField(charId: string, skillKey: string, genreIndex: number, field: 'job' | 'hobby' | 'perm' | 'temp', delta: number): Promise<boolean>

  // Custom skills
  addCustomSkill(charId: string): Promise<CustomSkill | null>
  deleteCustomSkill(charId: string, index: number): Promise<boolean>
  updateCustomSkillGrow(charId: string, index: number, grow: boolean): Promise<boolean>
  updateCustomSkillName(charId: string, index: number, name: string): Promise<boolean>
  updateCustomSkillField(charId: string, index: number, field: 'job' | 'hobby' | 'perm' | 'temp', delta: number): Promise<boolean>
  getCustomSkills(charId: string): Promise<CustomSkill[]>

  // Memos
  getMemos(charId: string): Promise<Record<string, string>>
  setMemo(charId: string, id: string, value: string): Promise<boolean>

  // Profile
  getProfile(charId: string): Promise<{ name: string; ruby: string }>
  setProfileName(charId: string, name: string): Promise<boolean>
  setProfileRuby(charId: string, ruby: string): Promise<boolean>

  // Image gallery
  getImageGallery(charId: string): Promise<ImageGalleryState>
  getCurrentImage(charId: string): Promise<CharacterImage | null>
  addImage(charId: string, filename: string, data: string, contentType: string): Promise<CharacterImage>
  deleteCurrentImage(charId: string): Promise<boolean>
  pinCurrentImage(charId: string): Promise<boolean>
  navigateImage(charId: string, direction: 'prev' | 'next'): Promise<boolean>
  setCurrentImageIndex(charId: string, index: number): Promise<boolean>
}
