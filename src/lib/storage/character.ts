/**
 * Character storage service.
 * Bridges the coalesce store with the application's data model.
 */

import type { CoalesceStore, CharacterData } from './coalesce'
import type { 
  StatusState, SkillsState, StatusVariable, StatusParameter, 
  SkillCategory, Skill, CustomSkill, SkillPoints, CharacterImage, ImageGalleryState
} from '../types'

// Default CoC 6e variables with min/max constraints
// Defaults are the expected values for each dice roll:
//   - 3D6 (STR, CON, POW, DEX, APP): expected = 10.5, use 11
//   - 2D6+6 (SIZ, INT): expected = 13
//   - 3D6+3 (EDU): expected = 13.5, use 14
const DEFAULT_VARIABLES: StatusVariable[] = [
  { key: 'STR', base: 11, perm: 0, temp: 0, min: 3, max: 18 },
  { key: 'CON', base: 11, perm: 0, temp: 0, min: 3, max: 18 },
  { key: 'POW', base: 11, perm: 0, temp: 0, min: 3, max: 18 },
  { key: 'DEX', base: 11, perm: 0, temp: 0, min: 3, max: 18 },
  { key: 'APP', base: 11, perm: 0, temp: 0, min: 3, max: 18 },
  { key: 'SIZ', base: 13, perm: 0, temp: 0, min: 8, max: 18 },
  { key: 'INT', base: 13, perm: 0, temp: 0, min: 8, max: 18 },
  { key: 'EDU', base: 14, perm: 0, temp: 0, min: 6, max: 21 },
]

// Skill definitions for CoC 6e
interface SkillDef {
  key: string
  category: string
  init: number | 'DEX*2' | 'EDU*5'
  order: number
  essential: boolean
  multi?: boolean
}

const SKILL_DEFS: SkillDef[] = [
  // 戦闘技能
  { key: '回避', category: '戦闘技能', init: 'DEX*2', order: 0, essential: true },
  { key: 'キック', category: '戦闘技能', init: 25, order: 1, essential: true },
  { key: '組み付き', category: '戦闘技能', init: 25, order: 2, essential: true },
  { key: 'こぶし', category: '戦闘技能', init: 50, order: 3, essential: true },
  { key: '頭突き', category: '戦闘技能', init: 10, order: 4, essential: true },
  { key: '投擲', category: '戦闘技能', init: 25, order: 5, essential: false },
  { key: 'マーシャルアーツ', category: '戦闘技能', init: 1, order: 6, essential: false },
  { key: '拳銃', category: '戦闘技能', init: 20, order: 7, essential: false },
  { key: 'サブマシンガン', category: '戦闘技能', init: 15, order: 8, essential: false },
  { key: 'ショットガン', category: '戦闘技能', init: 30, order: 9, essential: false },
  { key: 'マシンガン', category: '戦闘技能', init: 15, order: 10, essential: false },
  { key: 'ライフル', category: '戦闘技能', init: 25, order: 11, essential: false },
  // 探索技能
  { key: '応急手当', category: '探索技能', init: 30, order: 0, essential: true },
  { key: '鍵開け', category: '探索技能', init: 1, order: 1, essential: false },
  { key: '隠す', category: '探索技能', init: 15, order: 2, essential: false },
  { key: '隠れる', category: '探索技能', init: 10, order: 3, essential: false },
  { key: '聞き耳', category: '探索技能', init: 25, order: 4, essential: true },
  { key: '忍び歩き', category: '探索技能', init: 10, order: 5, essential: false },
  { key: '写真術', category: '探索技能', init: 10, order: 6, essential: false },
  { key: '精神分析', category: '探索技能', init: 1, order: 7, essential: false },
  { key: '追跡', category: '探索技能', init: 10, order: 8, essential: false },
  { key: '登攀', category: '探索技能', init: 40, order: 9, essential: false },
  { key: '図書館', category: '探索技能', init: 25, order: 10, essential: true },
  { key: '目星', category: '探索技能', init: 25, order: 11, essential: true },
  // 行動技能
  { key: '水泳', category: '行動技能', init: 25, order: 0, essential: false },
  { key: '製作', category: '行動技能', init: 5, order: 1, essential: false, multi: true },
  { key: '運転', category: '行動技能', init: 20, order: 2, essential: false, multi: true },
  { key: '操縦', category: '行動技能', init: 1, order: 3, essential: false, multi: true },
  { key: '乗馬', category: '行動技能', init: 5, order: 4, essential: false },
  { key: '跳躍', category: '行動技能', init: 25, order: 5, essential: false },
  { key: '電気修理', category: '行動技能', init: 10, order: 6, essential: false },
  { key: 'ナビゲート', category: '行動技能', init: 10, order: 7, essential: false },
  { key: '芸術', category: '行動技能', init: 5, order: 8, essential: false, multi: true },
  { key: '変装', category: '行動技能', init: 1, order: 9, essential: false },
  { key: '機械修理', category: '行動技能', init: 20, order: 10, essential: false },
  { key: '重機械操作', category: '行動技能', init: 1, order: 11, essential: false },
  // 交渉技能
  { key: '言いくるめ', category: '交渉技能', init: 5, order: 0, essential: false },
  { key: '信用', category: '交渉技能', init: 15, order: 1, essential: false },
  { key: '説得', category: '交渉技能', init: 15, order: 2, essential: false },
  { key: '値切り', category: '交渉技能', init: 5, order: 3, essential: false },
  { key: '母国語', category: '交渉技能', init: 'EDU*5', order: 4, essential: true, multi: true },
  { key: 'ほかの言語', category: '交渉技能', init: 1, order: 5, essential: false, multi: true },
  // 知識技能
  { key: '医学', category: '知識技能', init: 5, order: 0, essential: false },
  { key: 'オカルト', category: '知識技能', init: 5, order: 1, essential: false },
  { key: '化学', category: '知識技能', init: 1, order: 2, essential: false },
  { key: 'クトゥルフ神話', category: '知識技能', init: 0, order: 3, essential: false },
  { key: '経理', category: '知識技能', init: 10, order: 4, essential: false },
  { key: '考古学', category: '知識技能', init: 1, order: 5, essential: false },
  { key: 'コンピューター', category: '知識技能', init: 1, order: 6, essential: false },
  { key: '心理学', category: '知識技能', init: 5, order: 7, essential: false },
  { key: '人類学', category: '知識技能', init: 1, order: 8, essential: false },
  { key: '生物学', category: '知識技能', init: 1, order: 9, essential: false },
  { key: '地質学', category: '知識技能', init: 1, order: 10, essential: false },
  { key: '電子工学', category: '知識技能', init: 1, order: 11, essential: false },
  { key: '天文学', category: '知識技能', init: 1, order: 12, essential: false },
  { key: '博物学', category: '知識技能', init: 10, order: 13, essential: false },
  { key: '物理学', category: '知識技能', init: 1, order: 14, essential: false },
  { key: '法律', category: '知識技能', init: 5, order: 15, essential: false },
  { key: '薬学', category: '知識技能', init: 1, order: 16, essential: false },
  { key: '歴史', category: '知識技能', init: 20, order: 17, essential: false },
]

/**
 * Character service that wraps the coalesce store.
 */
export class CharacterService {
  constructor(private store: CoalesceStore) {}

  /**
   * Get or create a character, returning the hydrated data.
   */
  async getOrCreate(charId: string): Promise<CharacterData> {
    let data = await this.store.read(charId)
    if (!data) {
      data = await this.store.create(charId)
    }
    return data
  }

  /**
   * Get status state from character data.
   */
  getStatus(data: CharacterData, extra: { job: number; hobby: number }): StatusState {
    // Hydrate variables with min/max
    const variables: StatusVariable[] = DEFAULT_VARIABLES.map(def => {
      const stored = data.status.variables.find(v => v.key === def.key)
      return {
        ...def,
        base: stored?.base ?? def.base,
        perm: stored?.perm ?? 0,
        temp: stored?.temp ?? 0,
      }
    })

    const get = (key: string) => variables.find(v => v.key === key)
    const sum = (v: StatusVariable | undefined) => v ? v.base + v.perm + v.temp : 0

    const str = sum(get('STR'))
    const con = sum(get('CON'))
    const siz = sum(get('SIZ'))
    const pow = sum(get('POW'))
    const edu = sum(get('EDU'))
    const int = sum(get('INT'))

    // Compute derived values
    const sanDefault = pow * 5
    const ideaDefault = int * 5
    const luckDefault = pow * 5
    const knowDefault = edu * 5
    const totalJob = edu * 20 + extra.job
    const totalHobby = int * 10 + extra.hobby

    // Damage bonus calculation (matches Go implementation)
    const strSiz = str + siz
    let damageBonus: string
    if (strSiz < 13) damageBonus = '-1d6'
    else if (strSiz < 17) damageBonus = '-1d4'
    else if (strSiz < 25) damageBonus = '+0'
    else if (strSiz < 33) damageBonus = '+1d4'
    else damageBonus = '+1d6'

    // Hydrate parameters
    const parameters: StatusParameter[] = [
      { key: 'HP', value: data.status.parameters.find(p => p.key === 'HP')?.value ?? null, defaultValue: Math.ceil((con + siz) / 2) },
      { key: 'MP', value: data.status.parameters.find(p => p.key === 'MP')?.value ?? null, defaultValue: pow },
      { key: 'SAN', value: data.status.parameters.find(p => p.key === 'SAN')?.value ?? null, defaultValue: sanDefault },
    ]

    return {
      variables,
      computed: [
        { key: '初期SAN', value: sanDefault },
        { key: 'アイデア', value: ideaDefault },
        { key: '幸運', value: luckDefault },
        { key: '知識', value: knowDefault },
        { key: '職業P', value: totalJob },
        { key: '興味P', value: totalHobby },
      ],
      parameters,
      damageBonus: data.status.damageBonus ?? damageBonus,
    }
  }

  /**
   * Get skills state from character data.
   */
  getSkills(data: CharacterData, variables: StatusVariable[]): SkillsState {
    const get = (key: string) => variables.find(v => v.key === key)
    const sum = (v: StatusVariable | undefined) => v ? v.base + v.perm + v.temp : 0
    const dex = sum(get('DEX'))
    const edu = sum(get('EDU'))
    const int = sum(get('INT'))

    // Build skill categories
    const categoryMap = new Map<string, Skill[]>()
    
    for (const def of SKILL_DEFS) {
      // Calculate init value
      let init: number
      if (def.init === 'DEX*2') init = dex * 2
      else if (def.init === 'EDU*5') init = edu * 5
      else init = def.init

      const skill: Skill = {
        key: def.key,
        category: def.category,
        init,
        order: def.order,
        essential: def.essential,
      }

      if (def.multi) {
        // Multi-genre skill
        const stored = data.skills.multi[def.key]
        skill.multi = {
          genres: stored?.genres ?? [{ label: '', job: 0, hobby: 0, perm: 0, temp: 0, grow: false }]
        }
      } else {
        // Single skill
        const stored = data.skills.single[def.key]
        skill.single = stored ?? { job: 0, hobby: 0, perm: 0, temp: 0, grow: false }
      }

      if (!categoryMap.has(def.category)) {
        categoryMap.set(def.category, [])
      }
      categoryMap.get(def.category)!.push(skill)
    }

    // Convert to categories array
    const categoryOrder = ['戦闘技能', '探索技能', '行動技能', '交渉技能', '知識技能']
    const categories: SkillCategory[] = categoryOrder.map(name => ({
      name,
      skills: (categoryMap.get(name) || []).sort((a, b) => a.order - b.order)
    }))

    // Calculate consumed points
    let consumedJob = 0
    let consumedHobby = 0
    for (const cat of categories) {
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
    for (const custom of data.skills.custom) {
      consumedJob += custom.job
      consumedHobby += custom.hobby
    }

    const extra = data.skills.extra
    const totalJob = edu * 20 + extra.job
    const totalHobby = int * 10 + extra.hobby

    return {
      categories,
      customSkills: data.skills.custom,
      extra,
      remaining: {
        job: totalJob - consumedJob,
        hobby: totalHobby - consumedHobby,
      },
    }
  }

  /**
   * Get image gallery state.
   */
  getImageGallery(data: CharacterData): ImageGalleryState {
    return {
      images: data.images,
      currentIndex: data.currentImageIndex,
    }
  }

  /**
   * Write a value to the store.
   */
  async write(charId: string, path: string, value: unknown): Promise<void> {
    await this.store.write(charId, path, value)
  }
}
