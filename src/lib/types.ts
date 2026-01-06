// PageContext holds rendering context for the character sheet
export interface PageContext {
  isOwner: boolean
  preview: boolean
  memos: Record<string, string>
  basePath: string
}

export function isReadOnly(pc: PageContext): boolean {
  return !pc.isOwner || pc.preview
}

export function newPageContext(): PageContext {
  return {
    isOwner: true, // Hard-coded for development
    preview: false,
    memos: {},
    basePath: '',
  }
}

export function api(pc: PageContext, path: string): string {
  return pc.basePath + path
}

export function getMemo(pc: PageContext, id: string): string {
  return pc.memos[id] || ''
}

// StatusVariable represents a single ability score
export interface StatusVariable {
  key: string
  base: number
  perm: number
  temp: number
  min: number
  max: number
}

export function variableSum(v: StatusVariable): number {
  return v.base + v.perm + v.temp
}

// ComputedValue represents a derived value
export interface ComputedValue {
  key: string
  value: number
}

// StatusParameter represents an editable parameter
export interface StatusParameter {
  key: string
  value: number | null
  defaultValue: number
}

export function effectiveValue(p: StatusParameter): number {
  return p.value ?? p.defaultValue
}

// SkillPoints represents remaining skill points
export interface SkillPoints {
  job: number
  hobby: number
}

// StatusState holds all status data
export interface StatusState {
  variables: StatusVariable[]
  computed: ComputedValue[]
  parameters: StatusParameter[]
  damageBonus: string
}

// SingleSkillData holds point allocations for a single skill
export interface SingleSkillData {
  job: number
  hobby: number
  perm: number
  temp: number
  grow: boolean
}

// SkillGenre represents one specialty within a multi-genre skill
export interface SkillGenre {
  label: string
  job: number
  hobby: number
  perm: number
  temp: number
  grow: boolean
}

// Skill represents a skill
export interface Skill {
  key: string
  category: string
  init: number
  order: number
  essential: boolean
  single?: SingleSkillData
  multi?: { genres: SkillGenre[] }
}

export function skillTotal(s: Skill): number {
  if (s.single) {
    return s.init + s.single.job + s.single.hobby + s.single.perm + s.single.temp
  }
  return s.init
}

export function genreTotal(init: number, genre: SkillGenre): number {
  return init + genre.job + genre.hobby + genre.perm + genre.temp
}

// SkillCategory represents a group of skills
export interface SkillCategory {
  name: string
  skills: Skill[]
}

// CustomSkill represents a user-defined skill
export interface CustomSkill {
  name: string
  job: number
  hobby: number
  perm: number
  temp: number
  grow: boolean
}

// SkillsState holds all skills data
export interface SkillsState {
  categories: SkillCategory[]
  customSkills: CustomSkill[]
  extra: { job: number; hobby: number }
  remaining: SkillPoints
}

// SheetState holds all data needed to render a character sheet
export interface SheetState {
  pc: PageContext
  status: StatusState
  skills: SkillsState
}

// Image represents a character image
export interface CharacterImage {
  id: string
  filename: string
  data: string // Base64 encoded for in-memory, URL for R2
  contentType: string
  pinned: boolean
}

// ImageGalleryState holds image gallery data
export interface ImageGalleryState {
  images: CharacterImage[]
  currentIndex: number
}
