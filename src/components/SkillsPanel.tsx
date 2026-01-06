import type { FC } from 'hono/jsx'
import { NumberInput } from './NumberInput'
import { IconSquare, IconSquareCheck, IconCaretDown, IconPlus, IconTrash } from './icons'
import type { PageContext, SheetState, SkillsState, SkillCategory, Skill, SkillPoints, SingleSkillData, SkillGenre, CustomSkill } from '../lib/types'
import { isReadOnly, skillTotal, genreTotal } from '../lib/types'

type SkillsPanelProps = {
  state: SheetState
  oob?: boolean
}

// Extra points form
const ExtraPointsForm: FC<{ pc: PageContext; extra: { job: number; hobby: number }; oob?: boolean }> = ({ pc, extra, oob }) => {
  const readonly = isReadOnly(pc)
  
  return (
    <div class="flex flex-col gap-2 p-3 bg-slate-50 rounded-md" id="extra-points" {...(oob ? { 'hx-swap-oob': 'true' } : {})}>
      <div class="text-sm font-semibold text-slate-700">追加技能ポイント</div>
      <div class="grid grid-cols-[max-content_1fr] gap-2 items-center">
        <span class="text-sm font-medium text-slate-600">職業P</span>
        <div class="border border-slate-200 rounded-md bg-white min-w-0 overflow-hidden" id="extra-job-wrapper">
          <NumberInput
            id="extra-job"
            name="extra_job"
            value={extra.job}
            min={0}
            placeholder="0"
            readonly={readonly}
            basePath={pc.basePath}
            hxPost="/api/status/extra-job/adjust"
            hxTarget="#extra-job-wrapper"
            hxSwap="none"
          />
        </div>
        <span class="text-sm font-medium text-slate-600">興味P</span>
        <div class="border border-slate-200 rounded-md bg-white min-w-0 overflow-hidden" id="extra-hobby-wrapper">
          <NumberInput
            id="extra-hobby"
            name="extra_hobby"
            value={extra.hobby}
            min={0}
            placeholder="0"
            readonly={readonly}
            basePath={pc.basePath}
            hxPost="/api/status/extra-hobby/adjust"
            hxTarget="#extra-hobby-wrapper"
            hxSwap="none"
          />
        </div>
      </div>
    </div>
  )
}

// Skill grow button
const SkillGrowButton: FC<{ pc: PageContext; skillKey: string; grow: boolean; oob?: boolean }> = ({ pc, skillKey, grow, oob }) => {
  const readonly = isReadOnly(pc)
  
  return (
    <button
      type="button"
      id={`skill-grow-${skillKey}`}
      class={`w-8 h-8 inline-flex items-center justify-center border-none bg-transparent rounded-md cursor-pointer transition-colors duration-150 hover:bg-slate-100 ${grow ? 'text-blue-600' : 'text-slate-400'}`}
      title="成長チェック"
      disabled={readonly}
      hx-post={`${pc.basePath}/api/skill/${skillKey}/grow`}
      hx-swap="none"
      onclick="event.stopPropagation()"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      {grow ? <IconSquareCheck /> : <IconSquare />}
    </button>
  )
}

// Points breakdown display
const PointsBreakdown: FC<{ skillKey: string; data: SingleSkillData; oob?: boolean }> = ({ skillKey, data, oob }) => {
  const zeroClass = (v: number) => v === 0 ? 'text-slate-300' : ''
  const permTemp = data.perm + data.temp
  
  return (
    <div 
      id={`skill-breakdown-${skillKey}`}
      class="flex items-center h-8 text-xs font-mono text-slate-500" 
      title="職業P / 興味P / 増加分+一時的"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      <span class={`px-0.5 ${zeroClass(data.job)}`}>{String(data.job).padStart(2, '0')}</span>
      <span>/</span>
      <span class={`px-0.5 ${zeroClass(data.hobby)}`}>{String(data.hobby).padStart(2, '0')}</span>
      <span>/</span>
      <span class={`px-0.5 ${zeroClass(permTemp)}`}>{String(permTemp).padStart(2, '0')}</span>
    </div>
  )
}

// Skill name display (changes styling based on active/inactive)
const SkillName: FC<{ skill: Skill; oob?: boolean }> = ({ skill, oob }) => {
  const data = skill.single!
  const allocated = data.job + data.hobby + data.perm + data.temp
  const inactive = allocated === 0 && !data.grow
  
  return (
    <div 
      id={`skill-name-${skill.key}`}
      class={`flex-1 min-w-0 truncate ${skill.essential ? 'font-bold' : 'font-semibold'} ${inactive ? (skill.essential ? 'text-slate-600' : 'text-slate-400') : 'text-slate-900'}`}
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      {skill.key}
    </div>
  )
}

// Skill total display
const SkillTotal: FC<{ skill: Skill; oob?: boolean }> = ({ skill, oob }) => {
  return (
    <div 
      id={`skill-total-${skill.key}`}
      class="w-6 text-center font-semibold text-slate-800 tabular-nums" 
      title="達成値合計"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      {skillTotal(skill)}
    </div>
  )
}

// Single skill row (expandable)
const SkillRow: FC<{ pc: PageContext; skill: Skill; remaining: SkillPoints }> = ({ pc, skill, remaining }) => {
  const data = skill.single!
  const allocated = data.job + data.hobby + data.perm + data.temp
  const expanded = allocated > 0 || data.grow
  const readonly = isReadOnly(pc)
  
  return (
    <details class="skill-row" id={`skill-row-${skill.key}`} open={expanded}>
      <summary class="flex items-center gap-1 p-1 rounded-md cursor-pointer transition-colors duration-150 hover:bg-slate-50 list-none [&::-webkit-details-marker]:hidden">
        <SkillGrowButton pc={pc} skillKey={skill.key} grow={data.grow} />
        <SkillName skill={skill} />
        <PointsBreakdown skillKey={skill.key} data={data} />
        <SkillTotal skill={skill} />
        <div class="w-8 h-8 inline-flex items-center justify-center text-slate-400 transition-transform duration-150 [details[open]_&]:rotate-180">
          <IconCaretDown />
        </div>
      </summary>
      {!readonly && (
        <SkillDetail pc={pc} skill={skill} remaining={remaining} />
      )}
    </details>
  )
}

// Skill detail (expandable section)
const SkillDetail: FC<{ pc: PageContext; skill: Skill; remaining: SkillPoints; oob?: boolean }> = ({ pc, skill, remaining, oob }) => {
  const data = skill.single!
  const total = skillTotal(skill)
  const jobMax = Math.min(99 - (total - data.job), data.job + remaining.job)
  const hobbyMax = Math.min(99 - (total - data.hobby), data.hobby + remaining.hobby)
  const permMax = 99 - (total - data.perm)
  const tempMax = 99 - (total - data.temp)
  
  const inputConfig = (field: string, value: number, min: number, max: number) => ({
    id: `skill-${skill.key}-${field}`,
    name: `skill-${skill.key}-${field}`,
    value,
    min,
    max,
    readonly: false,
    basePath: pc.basePath,
    hxPost: `/api/skill/${skill.key}/${field}/adjust`,
    hxSwap: 'none', // Use OOB swaps instead
  })
  
  return (
    <div id={`skill-detail-${skill.key}`} class="grid grid-cols-[max-content_1fr] gap-2 items-center p-2 ml-8 bg-slate-50 rounded-md" {...(oob ? { 'hx-swap-oob': 'true' } : {})}>
      <span class="text-sm font-semibold text-slate-600">職業P</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('job', data.job, 0, jobMax)} />
      </div>
      <span class="text-sm font-semibold text-slate-600">興味P</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('hobby', data.hobby, 0, hobbyMax)} />
      </div>
      <span class="text-sm font-semibold text-slate-600">増加分</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('perm', data.perm, -(total - data.perm), permMax)} />
      </div>
      <span class="text-sm font-semibold text-slate-600">一時的</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('temp', data.temp, -(total - data.temp), tempMax)} />
      </div>
    </div>
  )
}

// ========== Multi-Genre Skill Components ==========

// Genre grow button
const GenreGrowButton: FC<{ pc: PageContext; skillKey: string; genreIndex: number; grow: boolean; oob?: boolean }> = ({ pc, skillKey, genreIndex, grow, oob }) => {
  const readonly = isReadOnly(pc)
  const id = `skill-genre-grow-${skillKey}-${genreIndex}`
  
  return (
    <button
      type="button"
      id={id}
      class={`w-8 h-8 inline-flex items-center justify-center border-none bg-transparent rounded-md cursor-pointer transition-colors duration-150 hover:bg-slate-100 ${grow ? 'text-blue-600' : 'text-slate-400'}`}
      title="成長チェック"
      disabled={readonly}
      hx-post={`${pc.basePath}/api/skill/${skillKey}/genre/${genreIndex}/grow`}
      hx-swap="none"
      onclick="event.stopPropagation()"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      {grow ? <IconSquareCheck /> : <IconSquare />}
    </button>
  )
}

// Genre points breakdown
const GenrePointsBreakdown: FC<{ skillKey: string; genreIndex: number; data: SkillGenre; oob?: boolean }> = ({ skillKey, genreIndex, data, oob }) => {
  const zeroClass = (v: number) => v === 0 ? 'text-slate-300' : ''
  const permTemp = data.perm + data.temp
  const id = `skill-genre-breakdown-${skillKey}-${genreIndex}`
  
  return (
    <div 
      id={id}
      class="flex items-center h-8 text-xs font-mono text-slate-500" 
      title="職業P / 興味P / 増加分+一時的"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      <span class={`px-0.5 ${zeroClass(data.job)}`}>{String(data.job).padStart(2, '0')}</span>
      <span>/</span>
      <span class={`px-0.5 ${zeroClass(data.hobby)}`}>{String(data.hobby).padStart(2, '0')}</span>
      <span>/</span>
      <span class={`px-0.5 ${zeroClass(permTemp)}`}>{String(permTemp).padStart(2, '0')}</span>
    </div>
  )
}

// Genre total display
const GenreTotal: FC<{ skillKey: string; genreIndex: number; init: number; genre: SkillGenre; oob?: boolean }> = ({ skillKey, genreIndex, init, genre, oob }) => {
  const id = `skill-genre-total-${skillKey}-${genreIndex}`
  return (
    <div 
      id={id}
      class="w-6 text-center font-semibold text-slate-800 tabular-nums" 
      title="達成値合計"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      {genreTotal(init, genre)}
    </div>
  )
}

// Genre name display (editable label)
const GenreName: FC<{ pc: PageContext; skillKey: string; genreIndex: number; label: string; genre: SkillGenre; oob?: boolean }> = ({ pc, skillKey, genreIndex, label, genre, oob }) => {
  const readonly = isReadOnly(pc)
  const allocated = genre.job + genre.hobby + genre.perm + genre.temp
  const inactive = allocated === 0 && !genre.grow
  const id = `skill-genre-name-${skillKey}-${genreIndex}`
  
  if (readonly) {
    return (
      <div 
        id={id}
        class={`flex-1 min-w-0 truncate font-semibold ${inactive ? 'text-slate-400' : 'text-slate-900'}`}
        {...(oob ? { 'hx-swap-oob': 'true' } : {})}
      >
        {label || '(未設定)'}
      </div>
    )
  }
  
  return (
    <input
      type="text"
      id={id}
      name={`genre_label_${skillKey}_${genreIndex}`}
      value={label}
      placeholder="(専門を入力)"
      class={`flex-1 min-w-0 w-full px-1 border-none bg-transparent text-sm font-semibold outline-none placeholder:text-slate-300 ${inactive ? 'text-slate-400' : 'text-slate-900'}`}
      hx-post={`${pc.basePath}/api/skill/${skillKey}/genre/${genreIndex}/label`}
      hx-trigger="input changed delay:500ms"
      hx-swap="none"
      onclick="event.stopPropagation()"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    />
  )
}

// Genre detail (expandable section)
const GenreDetail: FC<{ pc: PageContext; skillKey: string; genreIndex: number; init: number; genre: SkillGenre; remaining: SkillPoints; oob?: boolean }> = ({ pc, skillKey, genreIndex, init, genre, remaining, oob }) => {
  const total = genreTotal(init, genre)
  const jobMax = Math.min(99 - (total - genre.job), genre.job + remaining.job)
  const hobbyMax = Math.min(99 - (total - genre.hobby), genre.hobby + remaining.hobby)
  const permMax = 99 - (total - genre.perm)
  const tempMax = 99 - (total - genre.temp)
  const id = `skill-genre-detail-${skillKey}-${genreIndex}`
  
  const inputConfig = (field: string, value: number, min: number, max: number) => ({
    id: `skill-genre-${skillKey}-${genreIndex}-${field}`,
    name: `skill-genre-${skillKey}-${genreIndex}-${field}`,
    value,
    min,
    max,
    readonly: false,
    basePath: pc.basePath,
    hxPost: `/api/skill/${skillKey}/genre/${genreIndex}/${field}/adjust`,
    hxSwap: 'none',
  })
  
  return (
    <div id={id} class="grid grid-cols-[max-content_1fr] gap-2 items-center p-2 ml-8 bg-slate-50 rounded-md" {...(oob ? { 'hx-swap-oob': 'true' } : {})}>
      <span class="text-sm font-semibold text-slate-600">職業P</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('job', genre.job, 0, jobMax)} />
      </div>
      <span class="text-sm font-semibold text-slate-600">興味P</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('hobby', genre.hobby, 0, hobbyMax)} />
      </div>
      <span class="text-sm font-semibold text-slate-600">増加分</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('perm', genre.perm, -(total - genre.perm), permMax)} />
      </div>
      <span class="text-sm font-semibold text-slate-600">一時的</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('temp', genre.temp, -(total - genre.temp), tempMax)} />
      </div>
    </div>
  )
}

// Genre row (expandable)
const GenreRow: FC<{ pc: PageContext; skillKey: string; genreIndex: number; init: number; genre: SkillGenre; remaining: SkillPoints; showDelete: boolean }> = ({ pc, skillKey, genreIndex, init, genre, remaining, showDelete }) => {
  const allocated = genre.job + genre.hobby + genre.perm + genre.temp
  const expanded = allocated > 0 || genre.grow
  const readonly = isReadOnly(pc)
  const id = `skill-genre-row-${skillKey}-${genreIndex}`
  
  return (
    <details class="skill-genre-row" id={id} open={expanded}>
      <summary class="flex items-center gap-1 p-1 pl-4 rounded-md cursor-pointer transition-colors duration-150 hover:bg-slate-50 list-none [&::-webkit-details-marker]:hidden">
        <GenreGrowButton pc={pc} skillKey={skillKey} genreIndex={genreIndex} grow={genre.grow} />
        <GenreName pc={pc} skillKey={skillKey} genreIndex={genreIndex} label={genre.label} genre={genre} />
        <GenrePointsBreakdown skillKey={skillKey} genreIndex={genreIndex} data={genre} />
        <GenreTotal skillKey={skillKey} genreIndex={genreIndex} init={init} genre={genre} />
        {!readonly && showDelete && (
          <button
            type="button"
            class="w-8 h-8 inline-flex items-center justify-center border-none bg-transparent text-slate-400 rounded-md cursor-pointer transition-colors duration-150 hover:bg-red-50 hover:text-red-500"
            title="この専門を削除"
            hx-post={`${pc.basePath}/api/skill/${skillKey}/genre/${genreIndex}/delete`}
            hx-swap="none"
            hx-confirm="この専門を削除しますか？"
            onclick="event.stopPropagation()"
          >
            <IconTrash />
          </button>
        )}
        <div class="w-8 h-8 inline-flex items-center justify-center text-slate-400 transition-transform duration-150 [details[open]_&]:rotate-180">
          <IconCaretDown />
        </div>
      </summary>
      {!readonly && (
        <GenreDetail pc={pc} skillKey={skillKey} genreIndex={genreIndex} init={init} genre={genre} remaining={remaining} />
      )}
    </details>
  )
}

// Multi-genre skill (container for multiple genres)
const SkillMultiRow: FC<{ pc: PageContext; skill: Skill; remaining: SkillPoints }> = ({ pc, skill, remaining }) => {
  const genres = skill.multi!.genres
  const readonly = isReadOnly(pc)
  
  return (
    <div class="skill-multi" id={`skill-multi-${skill.key}`}>
      <div class="flex items-center gap-1 p-1 rounded-md">
        <div class="w-8 h-8"></div> {/* Spacer for alignment with single skills */}
        <div class={`flex-1 min-w-0 truncate font-semibold text-slate-700`}>
          {skill.key}
        </div>
        {!readonly && (
          <button
            type="button"
            class="w-8 h-8 inline-flex items-center justify-center border-none bg-transparent text-blue-500 rounded-md cursor-pointer transition-colors duration-150 hover:bg-blue-50"
            title="専門を追加"
            hx-post={`${pc.basePath}/api/skill/${skill.key}/genre/add`}
            hx-swap="none"
          >
            <IconPlus />
          </button>
        )}
      </div>
      <div class="flex flex-col gap-1 ml-2">
        {genres.map((genre, index) => (
          <GenreRow 
            pc={pc} 
            skillKey={skill.key} 
            genreIndex={index} 
            init={skill.init} 
            genre={genre} 
            remaining={remaining}
            showDelete={genres.length > 1}
          />
        ))}
        {!readonly && (
          <button
            type="button"
            class="w-full py-2 px-4 text-sm text-blue-600 border border-blue-300 rounded-md bg-white hover:bg-blue-50 transition-colors duration-150 cursor-pointer"
            hx-post={`${pc.basePath}/api/skill/${skill.key}/genre/add`}
            hx-swap="none"
          >
            「{skill.key}」技能を追加
          </button>
        )}
      </div>
    </div>
  )
}

// ========== Custom Skills Components ==========

// Custom skill total
function customSkillTotal(skill: CustomSkill): number {
  return skill.job + skill.hobby + skill.perm + skill.temp
}

// Custom skill grow button
const CustomSkillGrowButton: FC<{ pc: PageContext; index: number; grow: boolean; oob?: boolean }> = ({ pc, index, grow, oob }) => {
  const readonly = isReadOnly(pc)
  const id = `custom-skill-grow-${index}`
  
  return (
    <button
      type="button"
      id={id}
      class={`w-8 h-8 inline-flex items-center justify-center border-none bg-transparent rounded-md cursor-pointer transition-colors duration-150 hover:bg-slate-100 ${grow ? 'text-blue-600' : 'text-slate-400'}`}
      title="成長チェック"
      disabled={readonly}
      hx-post={`${pc.basePath}/api/skill/custom/${index}/grow`}
      hx-swap="none"
      onclick="event.stopPropagation()"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      {grow ? <IconSquareCheck /> : <IconSquare />}
    </button>
  )
}

// Custom skill points breakdown
const CustomSkillPointsBreakdown: FC<{ index: number; skill: CustomSkill; oob?: boolean }> = ({ index, skill, oob }) => {
  const zeroClass = (v: number) => v === 0 ? 'text-slate-300' : ''
  const permTemp = skill.perm + skill.temp
  const id = `custom-skill-breakdown-${index}`
  
  return (
    <div 
      id={id}
      class="flex items-center h-8 text-xs font-mono text-slate-500" 
      title="職業P / 興味P / 増加分+一時的"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      <span class={`px-0.5 ${zeroClass(skill.job)}`}>{String(skill.job).padStart(2, '0')}</span>
      <span>/</span>
      <span class={`px-0.5 ${zeroClass(skill.hobby)}`}>{String(skill.hobby).padStart(2, '0')}</span>
      <span>/</span>
      <span class={`px-0.5 ${zeroClass(permTemp)}`}>{String(permTemp).padStart(2, '0')}</span>
    </div>
  )
}

// Custom skill total display
const CustomSkillTotal: FC<{ index: number; skill: CustomSkill; oob?: boolean }> = ({ index, skill, oob }) => {
  const id = `custom-skill-total-${index}`
  return (
    <div 
      id={id}
      class="w-6 text-center font-semibold text-slate-800 tabular-nums" 
      title="達成値合計"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      {customSkillTotal(skill)}
    </div>
  )
}

// Custom skill name display (editable)
const CustomSkillName: FC<{ pc: PageContext; index: number; skill: CustomSkill; oob?: boolean }> = ({ pc, index, skill, oob }) => {
  const readonly = isReadOnly(pc)
  const allocated = skill.job + skill.hobby + skill.perm + skill.temp
  const inactive = allocated === 0 && !skill.grow
  const id = `custom-skill-name-${index}`
  
  if (readonly) {
    return (
      <div 
        id={id}
        class={`flex-1 min-w-0 truncate font-semibold ${inactive ? 'text-slate-400' : 'text-slate-900'}`}
        {...(oob ? { 'hx-swap-oob': 'true' } : {})}
      >
        {skill.name || '(未設定)'}
      </div>
    )
  }
  
  return (
    <input
      type="text"
      id={id}
      name={`custom_skill_name_${index}`}
      value={skill.name}
      placeholder="(技能名を入力)"
      class={`flex-1 min-w-0 w-full px-1 border-none bg-transparent text-sm font-semibold outline-none placeholder:text-slate-300 ${inactive ? 'text-slate-400' : 'text-slate-900'}`}
      hx-post={`${pc.basePath}/api/skill/custom/${index}/name`}
      hx-trigger="input changed delay:500ms"
      hx-swap="none"
      onclick="event.stopPropagation()"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    />
  )
}

// Custom skill detail (expandable section)
const CustomSkillDetail: FC<{ pc: PageContext; index: number; skill: CustomSkill; remaining: SkillPoints; oob?: boolean }> = ({ pc, index, skill, remaining, oob }) => {
  const total = customSkillTotal(skill)
  const jobMax = Math.min(99 - (total - skill.job), skill.job + remaining.job)
  const hobbyMax = Math.min(99 - (total - skill.hobby), skill.hobby + remaining.hobby)
  const permMax = 99 - (total - skill.perm)
  const tempMax = 99 - (total - skill.temp)
  const id = `custom-skill-detail-${index}`
  
  const inputConfig = (field: string, value: number, min: number, max: number) => ({
    id: `custom-skill-${index}-${field}`,
    name: `custom-skill-${index}-${field}`,
    value,
    min,
    max,
    readonly: false,
    basePath: pc.basePath,
    hxPost: `/api/skill/custom/${index}/${field}/adjust`,
    hxSwap: 'none',
  })
  
  return (
    <div id={id} class="grid grid-cols-[max-content_1fr] gap-2 items-center p-2 ml-8 bg-slate-50 rounded-md" {...(oob ? { 'hx-swap-oob': 'true' } : {})}>
      <span class="text-sm font-semibold text-slate-600">職業P</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('job', skill.job, 0, jobMax)} />
      </div>
      <span class="text-sm font-semibold text-slate-600">興味P</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('hobby', skill.hobby, 0, hobbyMax)} />
      </div>
      <span class="text-sm font-semibold text-slate-600">増加分</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('perm', skill.perm, -(total - skill.perm), permMax)} />
      </div>
      <span class="text-sm font-semibold text-slate-600">一時的</span>
      <div class="border border-slate-200 rounded-md bg-white">
        <NumberInput {...inputConfig('temp', skill.temp, -(total - skill.temp), tempMax)} />
      </div>
    </div>
  )
}

// Custom skill row
const CustomSkillRow: FC<{ pc: PageContext; index: number; skill: CustomSkill; remaining: SkillPoints }> = ({ pc, index, skill, remaining }) => {
  const allocated = skill.job + skill.hobby + skill.perm + skill.temp
  const expanded = allocated > 0 || skill.grow
  const readonly = isReadOnly(pc)
  const id = `custom-skill-row-${index}`
  
  return (
    <details class="custom-skill-row" id={id} open={expanded}>
      <summary class="flex items-center gap-1 p-1 rounded-md cursor-pointer transition-colors duration-150 hover:bg-slate-50 list-none [&::-webkit-details-marker]:hidden">
        <CustomSkillGrowButton pc={pc} index={index} grow={skill.grow} />
        <CustomSkillName pc={pc} index={index} skill={skill} />
        <CustomSkillPointsBreakdown index={index} skill={skill} />
        <CustomSkillTotal index={index} skill={skill} />
        {!readonly && (
          <button
            type="button"
            class="w-8 h-8 inline-flex items-center justify-center border-none bg-transparent text-slate-400 rounded-md cursor-pointer transition-colors duration-150 hover:bg-red-50 hover:text-red-500"
            title="この技能を削除"
            hx-post={`${pc.basePath}/api/skill/custom/${index}/delete`}
            hx-swap="none"
            hx-confirm="この技能を削除しますか？"
            onclick="event.stopPropagation()"
          >
            <IconTrash />
          </button>
        )}
        <div class="w-8 h-8 inline-flex items-center justify-center text-slate-400 transition-transform duration-150 [details[open]_&]:rotate-180">
          <IconCaretDown />
        </div>
      </summary>
      {!readonly && (
        <CustomSkillDetail pc={pc} index={index} skill={skill} remaining={remaining} />
      )}
    </details>
  )
}

// Custom skills section
const CustomSkillsSection: FC<{ pc: PageContext; customSkills: CustomSkill[]; remaining: SkillPoints; oob?: boolean }> = ({ pc, customSkills, remaining, oob }) => {
  const readonly = isReadOnly(pc)
  
  return (
    <div id="custom-skills-section" class="flex flex-col gap-2" {...(oob ? { 'hx-swap-oob': 'true' } : {})}>
      <div class="flex items-center justify-between py-2 border-b border-slate-200 mb-1">
        <h3 class="text-sm font-semibold text-slate-500 uppercase tracking-wider">独自技能</h3>
        {!readonly && (
          <button
            type="button"
            class="w-8 h-8 inline-flex items-center justify-center border-none bg-transparent text-blue-500 rounded-md cursor-pointer transition-colors duration-150 hover:bg-blue-50"
            title="独自技能を追加"
            hx-post={`${pc.basePath}/api/skill/custom/add`}
            hx-swap="none"
          >
            <IconPlus />
          </button>
        )}
      </div>
      {customSkills.length === 0 && (
        <div class="text-sm text-slate-400 text-center py-2">独自技能なし</div>
      )}
      <div class="flex flex-col gap-1">
        {customSkills.map((skill, index) => (
          <CustomSkillRow pc={pc} index={index} skill={skill} remaining={remaining} />
        ))}
      </div>
    </div>
  )
}

// Skill category block
const SkillCategoryBlock: FC<{ pc: PageContext; category: SkillCategory; remaining: SkillPoints }> = ({ pc, category, remaining }) => {
  return (
    <div class="flex flex-col gap-2">
      <h3 class="text-sm font-semibold text-slate-500 uppercase tracking-wider py-2 border-b border-slate-200 mb-1">
        {category.name}
      </h3>
      <div class="flex flex-col gap-1">
        {category.skills.map((skill) => {
          if (skill.single) {
            return <SkillRow pc={pc} skill={skill} remaining={remaining} />
          }
          if (skill.multi) {
            return <SkillMultiRow pc={pc} skill={skill} remaining={remaining} />
          }
          return null
        })}
      </div>
    </div>
  )
}

// Points display (floating badge)
export const PointsDisplay: FC<{ remaining: SkillPoints; oob?: boolean }> = ({ remaining, oob }) => {
  const jobClass = remaining.job < 0 ? 'text-red-600' : remaining.job > 0 ? 'text-green-600' : 'text-slate-800'
  const hobbyClass = remaining.hobby < 0 ? 'text-red-600' : remaining.hobby > 0 ? 'text-green-600' : 'text-slate-800'

  return (
    <div 
      id="points-display" 
      class="fixed top-2 right-2 sm:top-auto sm:bottom-4 sm:right-4 flex gap-2 bg-white rounded-md px-2 py-1 sm:px-3 sm:py-2 shadow-lg text-xs sm:text-sm font-semibold z-[100]"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      <span class="text-slate-600">職業P</span>
      <span class={jobClass}>{remaining.job}</span>
      <span class="text-slate-600">興味P</span>
      <span class={hobbyClass}>{remaining.hobby}</span>
    </div>
  )
}

// Main skills panel
export const SkillsPanel: FC<SkillsPanelProps> = ({ state, oob }) => {
  const { pc, skills } = state
  
  return (
    <div id="skills-panel" {...(oob ? { 'hx-swap-oob': 'true' } : {})}>
      <div class="bg-white rounded-lg p-4 shadow flex flex-col gap-4">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-semibold text-slate-800">技能</h2>
        </div>
        
        <div class="flex flex-col gap-4 2xl:grid 2xl:grid-cols-2">
          <div class="flex flex-col gap-4">
            <ExtraPointsForm pc={pc} extra={skills.extra} />
            {skills.categories.slice(0, 3).map((cat) => (
              <SkillCategoryBlock pc={pc} category={cat} remaining={skills.remaining} />
            ))}
          </div>
          <div class="flex flex-col gap-4">
            {skills.categories.slice(3).map((cat) => (
              <SkillCategoryBlock pc={pc} category={cat} remaining={skills.remaining} />
            ))}
            <CustomSkillsSection pc={pc} customSkills={skills.customSkills} remaining={skills.remaining} />
          </div>
        </div>
      </div>
    </div>
  )
}

// OOB Fragment: Skill update fragments (used when adjusting skill values)
// Returns granular updates instead of the full panel to preserve <details> state
export const SkillUpdateFragments: FC<{ state: SheetState; skillKey: string }> = ({ state, skillKey }) => {
  const skill = state.skills.categories.flatMap(c => c.skills).find(s => s.key === skillKey)
  if (!skill || !skill.single) return null
  
  return (
    <>
      <SkillDetail pc={state.pc} skill={skill} remaining={state.skills.remaining} oob />
      <PointsBreakdown skillKey={skill.key} data={skill.single} oob />
      <SkillTotal skill={skill} oob />
      <SkillName skill={skill} oob />
      <PointsDisplay remaining={state.skills.remaining} oob />
    </>
  )
}

// OOB Fragment: Skill grow update fragments
export const SkillGrowUpdateFragments: FC<{ state: SheetState; skillKey: string }> = ({ state, skillKey }) => {
  const skill = state.skills.categories.flatMap(c => c.skills).find(s => s.key === skillKey)
  if (!skill || !skill.single) return null
  
  return (
    <>
      <SkillGrowButton pc={state.pc} skillKey={skill.key} grow={skill.single.grow} oob />
      <SkillName skill={skill} oob />
    </>
  )
}

// OOB Fragment: Extra points update
export const ExtraPointsUpdateFragments: FC<{ state: SheetState }> = ({ state }) => {
  return (
    <>
      <ExtraPointsForm pc={state.pc} extra={state.skills.extra} oob />
      <PointsDisplay remaining={state.skills.remaining} oob />
    </>
  )
}

// ========== Multi-Genre Skill OOB Fragments ==========

// OOB Fragment: Genre update fragments (used when adjusting genre values)
export const GenreUpdateFragments: FC<{ state: SheetState; skillKey: string; genreIndex: number }> = ({ state, skillKey, genreIndex }) => {
  const skill = state.skills.categories.flatMap(c => c.skills).find(s => s.key === skillKey)
  if (!skill || !skill.multi) return null
  const genre = skill.multi.genres[genreIndex]
  if (!genre) return null
  
  return (
    <>
      <GenreDetail pc={state.pc} skillKey={skillKey} genreIndex={genreIndex} init={skill.init} genre={genre} remaining={state.skills.remaining} oob />
      <GenrePointsBreakdown skillKey={skillKey} genreIndex={genreIndex} data={genre} oob />
      <GenreTotal skillKey={skillKey} genreIndex={genreIndex} init={skill.init} genre={genre} oob />
      <GenreName pc={state.pc} skillKey={skillKey} genreIndex={genreIndex} label={genre.label} genre={genre} oob />
      <PointsDisplay remaining={state.skills.remaining} oob />
    </>
  )
}

// OOB Fragment: Genre grow update fragments
export const GenreGrowUpdateFragments: FC<{ state: SheetState; skillKey: string; genreIndex: number }> = ({ state, skillKey, genreIndex }) => {
  const skill = state.skills.categories.flatMap(c => c.skills).find(s => s.key === skillKey)
  if (!skill || !skill.multi) return null
  const genre = skill.multi.genres[genreIndex]
  if (!genre) return null
  
  return (
    <>
      <GenreGrowButton pc={state.pc} skillKey={skillKey} genreIndex={genreIndex} grow={genre.grow} oob />
      <GenreName pc={state.pc} skillKey={skillKey} genreIndex={genreIndex} label={genre.label} genre={genre} oob />
    </>
  )
}

// OOB Fragment: Multi-skill panel update (used when adding/deleting genres)
export const SkillMultiPanelFragment: FC<{ state: SheetState; skillKey: string }> = ({ state, skillKey }) => {
  const skill = state.skills.categories.flatMap(c => c.skills).find(s => s.key === skillKey)
  if (!skill || !skill.multi) return null
  
  return (
    <>
      <div id={`skill-multi-${skillKey}`} hx-swap-oob="true">
        <SkillMultiRow pc={state.pc} skill={skill} remaining={state.skills.remaining} />
      </div>
      <PointsDisplay remaining={state.skills.remaining} oob />
    </>
  )
}

// ========== Custom Skills OOB Fragments ==========

// OOB Fragment: Custom skill update fragments
export const CustomSkillUpdateFragments: FC<{ state: SheetState; index: number }> = ({ state, index }) => {
  const skill = state.skills.customSkills[index]
  if (!skill) return null
  
  return (
    <>
      <CustomSkillDetail pc={state.pc} index={index} skill={skill} remaining={state.skills.remaining} oob />
      <CustomSkillPointsBreakdown index={index} skill={skill} oob />
      <CustomSkillTotal index={index} skill={skill} oob />
      <CustomSkillName pc={state.pc} index={index} skill={skill} oob />
      <PointsDisplay remaining={state.skills.remaining} oob />
    </>
  )
}

// OOB Fragment: Custom skill grow update fragments
export const CustomSkillGrowUpdateFragments: FC<{ state: SheetState; index: number }> = ({ state, index }) => {
  const skill = state.skills.customSkills[index]
  if (!skill) return null
  
  return (
    <>
      <CustomSkillGrowButton pc={state.pc} index={index} grow={skill.grow} oob />
      <CustomSkillName pc={state.pc} index={index} skill={skill} oob />
    </>
  )
}

// OOB Fragment: Custom skills section update (used when adding/deleting custom skills)
export const CustomSkillsSectionFragment: FC<{ state: SheetState }> = ({ state }) => {
  return (
    <>
      <CustomSkillsSection pc={state.pc} customSkills={state.skills.customSkills} remaining={state.skills.remaining} oob />
      <PointsDisplay remaining={state.skills.remaining} oob />
    </>
  )
}
