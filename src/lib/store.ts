// In-memory store for development (will be replaced with D1/R2)
import type { StatusState, SkillsState, SkillPoints, StatusVariable, StatusParameter, Skill, SkillCategory, SingleSkillData, SkillGenre, CustomSkill, CharacterImage, ImageGalleryState } from './types'

// Default CoC 6e variables
// Defaults are the expected values for each dice roll:
//   - 3D6 (STR, CON, POW, DEX, APP): expected = 10.5, use 11
//   - 2D6+6 (SIZ, INT): expected = 13
//   - 3D6+3 (EDU): expected = 13.5, use 14
const defaultVariables: StatusVariable[] = [
  { key: 'STR', base: 11, perm: 0, temp: 0, min: 3, max: 18 },
  { key: 'CON', base: 11, perm: 0, temp: 0, min: 3, max: 18 },
  { key: 'POW', base: 11, perm: 0, temp: 0, min: 3, max: 18 },
  { key: 'DEX', base: 11, perm: 0, temp: 0, min: 3, max: 18 },
  { key: 'APP', base: 11, perm: 0, temp: 0, min: 3, max: 18 },
  { key: 'SIZ', base: 13, perm: 0, temp: 0, min: 8, max: 18 },
  { key: 'INT', base: 13, perm: 0, temp: 0, min: 8, max: 18 },
  { key: 'EDU', base: 14, perm: 0, temp: 0, min: 6, max: 21 },
]

// Calculate computed values from variables
function computeValues(vars: StatusVariable[], skillPoints: { total: SkillPoints }): { computed: { key: string; value: number }[]; damageBonus: string } {
  const get = (key: string) => vars.find(v => v.key === key)
  const sum = (v: StatusVariable | undefined) => v ? v.base + v.perm + v.temp : 0
  
  const str = sum(get('STR'))
  const con = sum(get('CON'))
  const siz = sum(get('SIZ'))
  const pow = sum(get('POW'))
  const dex = sum(get('DEX'))
  
  const sanDefault = pow * 5
  const ideaDefault = get('INT') ? sum(get('INT')) * 5 : 50
  const luckDefault = pow * 5
  const knowDefault = get('EDU') ? sum(get('EDU')) * 5 : 50
  
  // Damage bonus calculation (matches Go implementation)
  const strSiz = str + siz
  let damageBonus: string
  if (strSiz < 13) damageBonus = '-1d6'
  else if (strSiz < 17) damageBonus = '-1d4'
  else if (strSiz < 25) damageBonus = '+0'
  else if (strSiz < 33) damageBonus = '+1d4'
  else damageBonus = '+1d6'
  
  return {
    computed: [
      { key: '初期SAN', value: sanDefault },
      { key: 'アイデア', value: ideaDefault },
      { key: '幸運', value: luckDefault },
      { key: '知識', value: knowDefault },
      { key: '職業P', value: skillPoints.total.job },
      { key: '興味P', value: skillPoints.total.hobby },
    ],
    damageBonus,
  }
}

// Calculate default parameters from variables
function computeParameters(vars: StatusVariable[]): StatusParameter[] {
  const get = (key: string) => vars.find(v => v.key === key)
  const sum = (v: StatusVariable | undefined) => v ? v.base + v.perm + v.temp : 0
  
  const con = sum(get('CON'))
  const siz = sum(get('SIZ'))
  const pow = sum(get('POW'))
  
  return [
    { key: 'HP', value: null, defaultValue: Math.ceil((con + siz) / 2) },
    { key: 'MP', value: null, defaultValue: pow },
    { key: 'SAN', value: null, defaultValue: pow * 5 },
  ]
}

// Calculate skill points from variables
function computeSkillPoints(vars: StatusVariable[], extra: { job: number; hobby: number }): { total: SkillPoints; remaining: SkillPoints } {
  const get = (key: string) => vars.find(v => v.key === key)
  const sum = (v: StatusVariable | undefined) => v ? v.base + v.perm + v.temp : 0
  
  const edu = sum(get('EDU'))
  const int = sum(get('INT'))
  
  const totalJob = edu * 20 + extra.job
  const totalHobby = int * 10 + extra.hobby
  
  // TODO: Calculate consumed points from skills
  return {
    total: { job: totalJob, hobby: totalHobby },
    remaining: { job: totalJob, hobby: totalHobby },
  }
}

// Default skills for CoC 6e
function defaultSkills(vars: StatusVariable[]): SkillCategory[] {
  const get = (key: string) => vars.find(v => v.key === key)
  const sum = (v: StatusVariable | undefined) => v ? v.base + v.perm + v.temp : 0
  const dex = sum(get('DEX'))
  const edu = sum(get('EDU'))
  
  return [
    {
      name: '戦闘技能',
      skills: [
        { key: '回避', category: '戦闘技能', init: dex * 2, order: 0, essential: true, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'キック', category: '戦闘技能', init: 25, order: 1, essential: true, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '組み付き', category: '戦闘技能', init: 25, order: 2, essential: true, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'こぶし', category: '戦闘技能', init: 50, order: 3, essential: true, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '頭突き', category: '戦闘技能', init: 10, order: 4, essential: true, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '投擲', category: '戦闘技能', init: 25, order: 5, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'マーシャルアーツ', category: '戦闘技能', init: 1, order: 6, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '拳銃', category: '戦闘技能', init: 20, order: 7, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'サブマシンガン', category: '戦闘技能', init: 15, order: 8, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'ショットガン', category: '戦闘技能', init: 30, order: 9, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'マシンガン', category: '戦闘技能', init: 15, order: 10, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'ライフル', category: '戦闘技能', init: 25, order: 11, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
      ],
    },
    {
      name: '探索技能',
      skills: [
        { key: '応急手当', category: '探索技能', init: 30, order: 0, essential: true, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '鍵開け', category: '探索技能', init: 1, order: 1, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '隠す', category: '探索技能', init: 15, order: 2, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '隠れる', category: '探索技能', init: 10, order: 3, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '聞き耳', category: '探索技能', init: 25, order: 4, essential: true, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '忍び歩き', category: '探索技能', init: 10, order: 5, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '写真術', category: '探索技能', init: 10, order: 6, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '精神分析', category: '探索技能', init: 1, order: 7, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '追跡', category: '探索技能', init: 10, order: 8, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '登攀', category: '探索技能', init: 40, order: 9, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '図書館', category: '探索技能', init: 25, order: 10, essential: true, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '目星', category: '探索技能', init: 25, order: 11, essential: true, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
      ],
    },
    {
      name: '行動技能',
      skills: [
        { key: '水泳', category: '行動技能', init: 25, order: 0, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '製作', category: '行動技能', init: 5, order: 1, essential: false, multi: { genres: [{ label: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }] } },
        { key: '運転', category: '行動技能', init: 20, order: 2, essential: false, multi: { genres: [{ label: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }] } },
        { key: '操縦', category: '行動技能', init: 1, order: 3, essential: false, multi: { genres: [{ label: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }] } },
        { key: '乗馬', category: '行動技能', init: 5, order: 4, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '跳躍', category: '行動技能', init: 25, order: 5, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '電気修理', category: '行動技能', init: 10, order: 6, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'ナビゲート', category: '行動技能', init: 10, order: 7, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '芸術', category: '行動技能', init: 5, order: 8, essential: false, multi: { genres: [{ label: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }] } },
        { key: '変装', category: '行動技能', init: 1, order: 9, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '機械修理', category: '行動技能', init: 20, order: 10, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '重機械操作', category: '行動技能', init: 1, order: 11, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
      ],
    },
    {
      name: '交渉技能',
      skills: [
        { key: '言いくるめ', category: '交渉技能', init: 5, order: 0, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '信用', category: '交渉技能', init: 15, order: 1, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '説得', category: '交渉技能', init: 15, order: 2, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '値切り', category: '交渉技能', init: 5, order: 3, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '母国語', category: '交渉技能', init: edu * 5, order: 4, essential: true, multi: { genres: [{ label: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }] } },
        { key: 'ほかの言語', category: '交渉技能', init: 1, order: 5, essential: false, multi: { genres: [{ label: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }] } },
      ],
    },
    {
      name: '知識技能',
      skills: [
        { key: '医学', category: '知識技能', init: 5, order: 0, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'オカルト', category: '知識技能', init: 5, order: 1, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '化学', category: '知識技能', init: 1, order: 2, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'クトゥルフ神話', category: '知識技能', init: 0, order: 3, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '経理', category: '知識技能', init: 10, order: 4, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '考古学', category: '知識技能', init: 1, order: 5, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: 'コンピューター', category: '知識技能', init: 1, order: 6, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '心理学', category: '知識技能', init: 5, order: 7, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '人類学', category: '知識技能', init: 1, order: 8, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '生物学', category: '知識技能', init: 1, order: 9, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '地質学', category: '知識技能', init: 1, order: 10, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '電子工学', category: '知識技能', init: 1, order: 11, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '天文学', category: '知識技能', init: 1, order: 12, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '博物学', category: '知識技能', init: 10, order: 13, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '物理学', category: '知識技能', init: 1, order: 14, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '法律', category: '知識技能', init: 5, order: 15, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '薬学', category: '知識技能', init: 1, order: 16, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
        { key: '歴史', category: '知識技能', init: 20, order: 17, essential: false, single: { job: 0, hobby: 0, perm: 0, temp: 0, grow: false } },
      ],
    },
  ]
}

// Character data store (in-memory for now)
interface CharacterData {
  variables: StatusVariable[]
  parameters: StatusParameter[]
  memos: Record<string, string>
  skills: SkillCategory[]
  extra: { job: number; hobby: number }
  customSkills: CustomSkill[]
  profile: { name: string; ruby: string }
  damageBonus: string | null // Override for computed damage bonus
  images: CharacterImage[]
  currentImageIndex: number
}

const characters: Record<string, CharacterData> = {}

function getOrCreateCharacter(id: string): CharacterData {
  if (!characters[id]) {
    characters[id] = {
      variables: JSON.parse(JSON.stringify(defaultVariables)),
      parameters: computeParameters(defaultVariables),
      memos: {},
      skills: defaultSkills(defaultVariables),
      extra: { job: 0, hobby: 0 },
      customSkills: [],
      profile: { name: '', ruby: '' },
      damageBonus: null,
      images: [],
      currentImageIndex: 0,
    }
  }
  return characters[id]
}

export function getStatus(charId: string): StatusState {
  const char = getOrCreateCharacter(charId)
  const skillPoints = computeSkillPoints(char.variables, char.extra)
  const { computed, damageBonus: computedDB } = computeValues(char.variables, skillPoints)
  return {
    variables: char.variables,
    computed,
    parameters: char.parameters,
    damageBonus: char.damageBonus ?? computedDB,
  }
}

export function setVariableBase(charId: string, key: string, value: number): StatusVariable | null {
  const char = getOrCreateCharacter(charId)
  const variable = char.variables.find(v => v.key === key)
  if (!variable) return null
  
  // Clamp value
  const clamped = Math.max(variable.min, Math.min(variable.max, value))
  if (variable.base === clamped) return null
  
  variable.base = clamped
  return variable
}

export function getMemos(charId: string): Record<string, string> {
  const char = getOrCreateCharacter(charId)
  return char.memos
}

export function setMemo(charId: string, id: string, value: string): boolean {
  const char = getOrCreateCharacter(charId)
  if (char.memos[id] === value) return false
  char.memos[id] = value
  return true
}

export function getSkills(charId: string): SkillsState {
  const char = getOrCreateCharacter(charId)
  const { remaining } = computeSkillPoints(char.variables, char.extra)
  
  // Get current variable values for dynamic init calculation
  const get = (key: string) => char.variables.find(v => v.key === key)
  const sum = (v: StatusVariable | undefined) => v ? v.base + v.perm + v.temp : 0
  const dex = sum(get('DEX'))
  const edu = sum(get('EDU'))
  
  // Update dynamic init values
  for (const cat of char.skills) {
    for (const skill of cat.skills) {
      // 回避 depends on DEX*2
      if (skill.key === '回避') {
        skill.init = dex * 2
      }
      // 母国語 depends on EDU*5
      if (skill.key === '母国語') {
        skill.init = edu * 5
      }
    }
  }
  
  // Calculate consumed points
  let consumedJob = 0
  let consumedHobby = 0
  for (const cat of char.skills) {
    for (const skill of cat.skills) {
      if (skill.single) {
        consumedJob += skill.single.job
        consumedHobby += skill.single.hobby
      }
      if (skill.multi) {
        for (const g of skill.multi.genres) {
          consumedJob += g.job
          consumedHobby += g.hobby
        }
      }
    }
  }
  
  // Include custom skills in point calculation
  for (const custom of char.customSkills) {
    consumedJob += custom.job
    consumedHobby += custom.hobby
  }
  
  return {
    categories: char.skills,
    customSkills: char.customSkills,
    extra: char.extra,
    remaining: {
      job: remaining.job - consumedJob,
      hobby: remaining.hobby - consumedHobby,
    },
  }
}

export function getSkill(charId: string, key: string): Skill | null {
  const char = getOrCreateCharacter(charId)
  for (const cat of char.skills) {
    const skill = cat.skills.find(s => s.key === key)
    if (skill) return skill
  }
  return null
}

export function updateSkill(charId: string, key: string, updates: Partial<SingleSkillData>): Skill | null {
  const skill = getSkill(charId, key)
  if (!skill || !skill.single) return null
  
  Object.assign(skill.single, updates)
  return skill
}

export function setExtraJob(charId: string, value: number): boolean {
  const char = getOrCreateCharacter(charId)
  const newValue = Math.max(0, value)
  if (char.extra.job === newValue) return false
  char.extra.job = newValue
  return true
}

export function setExtraHobby(charId: string, value: number): boolean {
  const char = getOrCreateCharacter(charId)
  const newValue = Math.max(0, value)
  if (char.extra.hobby === newValue) return false
  char.extra.hobby = newValue
  return true
}

export function adjustExtraJob(charId: string, delta: number): boolean {
  const char = getOrCreateCharacter(charId)
  const newValue = Math.max(0, char.extra.job + delta)
  if (char.extra.job === newValue) return false
  char.extra.job = newValue
  return true
}

export function adjustExtraHobby(charId: string, delta: number): boolean {
  const char = getOrCreateCharacter(charId)
  const newValue = Math.max(0, char.extra.hobby + delta)
  if (char.extra.hobby === newValue) return false
  char.extra.hobby = newValue
  return true
}

// ========== Multi-Genre Skill Functions ==========

export function addGenre(charId: string, skillKey: string): SkillGenre | null {
  const skill = getSkill(charId, skillKey)
  if (!skill || !skill.multi) return null
  
  const newGenre: SkillGenre = { label: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }
  skill.multi.genres.push(newGenre)
  return newGenre
}

export function deleteGenre(charId: string, skillKey: string, genreIndex: number): boolean {
  const skill = getSkill(charId, skillKey)
  if (!skill || !skill.multi) return false
  if (skill.multi.genres.length <= 1) return false // Keep at least one
  if (genreIndex < 0 || genreIndex >= skill.multi.genres.length) return false
  
  skill.multi.genres.splice(genreIndex, 1)
  return true
}

export function updateGenreGrow(charId: string, skillKey: string, genreIndex: number, grow: boolean): boolean {
  const skill = getSkill(charId, skillKey)
  if (!skill || !skill.multi) return false
  if (genreIndex < 0 || genreIndex >= skill.multi.genres.length) return false
  
  skill.multi.genres[genreIndex].grow = grow
  return true
}

export function updateGenreLabel(charId: string, skillKey: string, genreIndex: number, label: string): boolean {
  const skill = getSkill(charId, skillKey)
  if (!skill || !skill.multi) return false
  if (genreIndex < 0 || genreIndex >= skill.multi.genres.length) return false
  
  skill.multi.genres[genreIndex].label = label
  return true
}

export function updateGenreField(charId: string, skillKey: string, genreIndex: number, field: 'job' | 'hobby' | 'perm' | 'temp', delta: number): boolean {
  const skill = getSkill(charId, skillKey)
  if (!skill || !skill.multi) return false
  if (genreIndex < 0 || genreIndex >= skill.multi.genres.length) return false
  
  const genre = skill.multi.genres[genreIndex]
  const newValue = Math.max(0, genre[field] + delta)
  genre[field] = newValue
  return true
}

// ========== Custom Skill Functions ==========

export function addCustomSkill(charId: string): CustomSkill | null {
  const char = getOrCreateCharacter(charId)
  const newSkill: CustomSkill = { name: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }
  char.customSkills.push(newSkill)
  return newSkill
}

export function deleteCustomSkill(charId: string, index: number): boolean {
  const char = getOrCreateCharacter(charId)
  if (index < 0 || index >= char.customSkills.length) return false
  char.customSkills.splice(index, 1)
  return true
}

export function updateCustomSkillGrow(charId: string, index: number, grow: boolean): boolean {
  const char = getOrCreateCharacter(charId)
  if (index < 0 || index >= char.customSkills.length) return false
  char.customSkills[index].grow = grow
  return true
}

export function updateCustomSkillName(charId: string, index: number, name: string): boolean {
  const char = getOrCreateCharacter(charId)
  if (index < 0 || index >= char.customSkills.length) return false
  char.customSkills[index].name = name
  return true
}

export function updateCustomSkillField(charId: string, index: number, field: 'job' | 'hobby' | 'perm' | 'temp', delta: number): boolean {
  const char = getOrCreateCharacter(charId)
  if (index < 0 || index >= char.customSkills.length) return false
  
  const skill = char.customSkills[index]
  const newValue = Math.max(0, skill[field] + delta)
  skill[field] = newValue
  return true
}

export function getCustomSkills(charId: string): CustomSkill[] {
  const char = getOrCreateCharacter(charId)
  return char.customSkills
}

// ========== Parameter Functions ==========

export function setParameter(charId: string, key: string, value: number): boolean {
  const char = getOrCreateCharacter(charId)
  const param = char.parameters.find(p => p.key === key)
  if (!param) return false
  
  const newValue = Math.max(0, value)
  param.value = newValue
  return true
}

export function adjustParameter(charId: string, key: string, delta: number): boolean {
  const char = getOrCreateCharacter(charId)
  const param = char.parameters.find(p => p.key === key)
  if (!param) return false
  
  const current = param.value ?? param.defaultValue
  const newValue = Math.max(0, current + delta)
  param.value = newValue
  return true
}

// ========== Profile Functions ==========

export function getProfile(charId: string): { name: string; ruby: string } {
  const char = getOrCreateCharacter(charId)
  return char.profile
}

export function setProfileName(charId: string, name: string): boolean {
  const char = getOrCreateCharacter(charId)
  if (char.profile.name === name) return false
  char.profile.name = name
  return true
}

export function setProfileRuby(charId: string, ruby: string): boolean {
  const char = getOrCreateCharacter(charId)
  if (char.profile.ruby === ruby) return false
  char.profile.ruby = ruby
  return true
}

// ========== Damage Bonus Function ==========

export function setDamageBonus(charId: string, value: string): boolean {
  const char = getOrCreateCharacter(charId)
  if (char.damageBonus === value) return false
  char.damageBonus = value
  return true
}

// ========== Random Stats Function ==========

function roll3d6(): number {
  return Math.floor(Math.random() * 6) + 1 +
         Math.floor(Math.random() * 6) + 1 +
         Math.floor(Math.random() * 6) + 1
}

function roll2d6plus6(): number {
  return Math.floor(Math.random() * 6) + 1 +
         Math.floor(Math.random() * 6) + 1 + 6
}

function roll3d6plus3(): number {
  return roll3d6() + 3
}

export function randomizeStats(charId: string): boolean {
  const char = getOrCreateCharacter(charId)
  
  // CoC 6e dice rules:
  // STR, CON, POW, DEX, APP: 3d6
  // SIZ, INT: 2d6+6
  // EDU: 3d6+3
  for (const v of char.variables) {
    switch (v.key) {
      case 'STR':
      case 'CON':
      case 'POW':
      case 'DEX':
      case 'APP':
        v.base = roll3d6()
        break
      case 'SIZ':
      case 'INT':
        v.base = roll2d6plus6()
        break
      case 'EDU':
        v.base = roll3d6plus3()
        break
    }
  }
  
  return true
}

// ========== Image Gallery Functions ==========

function generateImageId(): string {
  return 'img_' + Date.now().toString(36) + Math.random().toString(36).substr(2, 9)
}

export function getImageGallery(charId: string): ImageGalleryState {
  const char = getOrCreateCharacter(charId)
  return {
    images: char.images,
    currentIndex: char.currentImageIndex,
  }
}

export function getCurrentImage(charId: string): CharacterImage | null {
  const char = getOrCreateCharacter(charId)
  if (char.images.length === 0) return null
  const index = Math.min(char.currentImageIndex, char.images.length - 1)
  return char.images[index]
}

export function addImage(charId: string, filename: string, data: string, contentType: string): CharacterImage {
  const char = getOrCreateCharacter(charId)
  const image: CharacterImage = {
    id: generateImageId(),
    filename,
    data,
    contentType,
    pinned: false,
  }
  char.images.push(image)
  char.currentImageIndex = char.images.length - 1
  return image
}

export function deleteImage(charId: string, imageId: string): boolean {
  const char = getOrCreateCharacter(charId)
  const index = char.images.findIndex(img => img.id === imageId)
  if (index === -1) return false
  
  char.images.splice(index, 1)
  
  // Adjust current index if needed
  if (char.images.length === 0) {
    char.currentImageIndex = 0
  } else if (char.currentImageIndex >= char.images.length) {
    char.currentImageIndex = char.images.length - 1
  }
  
  return true
}

export function deleteCurrentImage(charId: string): boolean {
  const char = getOrCreateCharacter(charId)
  if (char.images.length === 0) return false
  const currentImage = char.images[char.currentImageIndex]
  return deleteImage(charId, currentImage.id)
}

export function pinImage(charId: string, imageId: string): boolean {
  const char = getOrCreateCharacter(charId)
  const image = char.images.find(img => img.id === imageId)
  if (!image) return false
  
  // Unpin all other images
  for (const img of char.images) {
    img.pinned = img.id === imageId ? !img.pinned : false
  }
  
  return true
}

export function pinCurrentImage(charId: string): boolean {
  const char = getOrCreateCharacter(charId)
  if (char.images.length === 0) return false
  const currentImage = char.images[char.currentImageIndex]
  return pinImage(charId, currentImage.id)
}

export function navigateImage(charId: string, direction: 'prev' | 'next'): boolean {
  const char = getOrCreateCharacter(charId)
  if (char.images.length <= 1) return false
  
  if (direction === 'prev') {
    char.currentImageIndex = (char.currentImageIndex - 1 + char.images.length) % char.images.length
  } else {
    char.currentImageIndex = (char.currentImageIndex + 1) % char.images.length
  }
  
  return true
}

export function setCurrentImageIndex(charId: string, index: number): boolean {
  const char = getOrCreateCharacter(charId)
  if (index < 0 || index >= char.images.length) return false
  char.currentImageIndex = index
  return true
}

export function getPinnedImage(charId: string): CharacterImage | null {
  const char = getOrCreateCharacter(charId)
  return char.images.find(img => img.pinned) || null
}
