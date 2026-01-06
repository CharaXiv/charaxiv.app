import { Hono } from 'hono'
import { Layout } from './components/Layout'
import { Button } from './components/Button'
import { NumberInput } from './components/NumberInput'
import { Profile, ImageGallery, GalleryModal } from './components/Profile'
import { ScenarioMemoGroup, MemoGroup } from './components/MarkdownEditor'
import { SkillsPanel, PointsDisplay, SkillUpdateFragments, SkillGrowUpdateFragments, ExtraPointsUpdateFragments, GenreUpdateFragments, GenreGrowUpdateFragments, SkillMultiPanelFragment, CustomSkillUpdateFragments, CustomSkillGrowUpdateFragments, CustomSkillsSectionFragment } from './components/SkillsPanel'
import { IconEye, IconAddressBook, IconArrowUpFromBracket } from './components/icons'
import type { AsyncStore } from './lib/storage'
import { createMemoryStore, createD1R2Store } from './lib/storage'
import type { PageContext, SheetState, ImageGalleryState } from './lib/types'
import { isReadOnly, getMemo, api, variableSum, effectiveValue, skillTotal } from './lib/types'

// Static files - inlined at build time
import stylesCSS from './static-styles.txt'
import componentsJS from './static-components.txt'

type Bindings = {
  DB: D1Database
  BUCKET: R2Bucket
}

type Variables = {
  store: AsyncStore
}

const app = new Hono<{ Bindings: Bindings; Variables: Variables }>()

// Middleware: attach store to context
// Use D1/R2 when available (Workers), fall back to in-memory (Bun dev)
let memoryStore: AsyncStore | null = null

app.use('*', async (c, next) => {
  if (c.env?.DB && c.env?.BUCKET) {
    // Workers environment with D1/R2
    c.set('store', createD1R2Store(c.env.DB, c.env.BUCKET))
  } else {
    // Bun dev environment - use shared in-memory store
    if (!memoryStore) {
      memoryStore = createMemoryStore()
    }
    c.set('store', memoryStore)
  }
  await next()
})

app.get('/styles.css', (c) => {
  c.header('Content-Type', 'text/css')
  return c.body(stylesCSS)
})

app.get('/components.js', (c) => {
  c.header('Content-Type', 'application/javascript')
  return c.body(componentsJS)
})

// Health check
app.get('/health', (c) => c.text('ok'))

// Constants
const CHAR_ID = 'demo'
const BASE_PATH = '/cthulhu6'

// Helper to build page context (async due to memos)
async function buildPageContext(store: AsyncStore, preview = false): Promise<PageContext> {
  return {
    isOwner: true,
    preview,
    memos: await store.getMemos(CHAR_ID),
    basePath: BASE_PATH,
  }
}

// Helper to build sheet state (async)
async function buildSheetState(store: AsyncStore, preview = false): Promise<SheetState> {
  const [pc, status, skills] = await Promise.all([
    buildPageContext(store, preview),
    store.getStatus(CHAR_ID),
    store.getSkills(CHAR_ID),
  ])
  return { pc, status, skills }
}

// Header actions component
const SheetHeaderActions = ({ pc }: { pc: PageContext }) => (
  <div id="header-actions" class="flex items-center gap-2">
    {pc.isOwner && (
      pc.preview ? (
        <Button
          variant="solid-blue"
          icon
          title="Exit preview mode"
          hxPost={api(pc, '/api/preview/off')}
          hxTarget="#memo-group"
          hxSwap="outerHTML"
        >
          <IconEye />
        </Button>
      ) : (
        <Button
          variant="ghost"
          icon
          title="Preview as visitor"
          hxPost={api(pc, '/api/preview/on')}
          hxTarget="#memo-group"
          hxSwap="outerHTML"
        >
          <IconEye />
        </Button>
      )
    )}
    <Button variant="ghost-blue" icon href="/characters" title="Character page">
      <IconAddressBook />
    </Button>
    <Button variant="solid-blue" icon title="Share options">
      <IconArrowUpFromBracket />
    </Button>
  </div>
)

// Status panel component
const StatusPanel = ({ state, oob = false }: { state: SheetState; oob?: boolean }) => {
  const { pc, status } = state
  const readonly = isReadOnly(pc)

  return (
    <div id="status-panel" {...(oob ? { 'hx-swap-oob': 'true' } : {})}>
      <div class="bg-white rounded-lg p-4 shadow flex flex-col gap-4">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-semibold text-slate-800">能力値</h2>
          {!readonly && (
            <Button 
              variant="outline-blue" 
              title="ランダム"
              hxPost={api(pc, '/api/status/random')}
              hxSwap="none"
            >
              ランダム
            </Button>
          )}
        </div>
        
        {/* Variables grid */}
        <div class="grid grid-cols-[48px_1fr_40px] gap-2 items-center">
          {status.variables.map((v) => (
            <>
              <div class="font-semibold text-slate-700 text-sm tracking-wider">{v.key}</div>
              <div class="w-full border border-slate-200 rounded-md bg-white hover:border-slate-300 focus-within:border-blue-500 transition-colors duration-150">
                <NumberInput
                  id={`status-${v.key}`}
                  name={`status_${v.key}`}
                  value={v.base}
                  min={v.min}
                  max={v.max}
                  placeholder={v.key}
                  readonly={readonly}
                  basePath={pc.basePath}
                />
              </div>
              <div class="font-semibold text-slate-800 text-center">{variableSum(v)}</div>
            </>
          ))}
        </div>

        {/* Computed values - 2 columns, 3 rows */}
        <div class="grid grid-cols-[1fr_48px_1fr_48px] gap-2 items-center">
          {status.computed.map((c) => (
            <>
              <div class="font-semibold text-slate-600 text-sm text-right pr-2">{c.key}</div>
              <div class="font-semibold text-slate-800 text-center">{c.value}</div>
            </>
          ))}
        </div>

        <div class="h-px bg-slate-200 my-2"></div>

        {/* Parameters */}
        <h2 class="text-xl font-semibold text-slate-800">パラメーター</h2>
        <div class="grid grid-cols-[48px_1fr_40px] gap-2 items-center">
          {status.parameters.map((p) => (
            <>
              <div class="font-semibold text-slate-700 text-sm">{p.key}</div>
              <div class="w-full border border-slate-200 rounded-md bg-white hover:border-slate-300 focus-within:border-blue-500 transition-colors duration-150">
                <NumberInput
                  id={`param-${p.key}`}
                  name={`param_${p.key}`}
                  value={effectiveValue(p)}
                  min={0}
                  placeholder={String(p.defaultValue)}
                  readonly={readonly}
                  basePath={pc.basePath}
                  hxSwap="none"
                />
              </div>
              <div class="font-semibold text-slate-400 text-center text-sm">{p.defaultValue}</div>
            </>
          ))}
          
          {/* Damage bonus */}
          <div class="font-semibold text-slate-700 text-sm">DB</div>
          <input
            type="text"
            id="param-db"
            name="param_db"
            class="w-full h-8 px-2 border border-slate-200 rounded-md text-center font-semibold outline-none transition-colors duration-150 hover:border-slate-300 focus:border-blue-500"
            value={status.damageBonus}
            placeholder={status.damageBonus}
            readonly={readonly}
            hx-post={api(pc, '/api/param/db/set')}
            hx-trigger="input changed delay:1500ms"
            hx-swap="none"
          />
          <div></div>
          
          {/* Indefinite */}
          {(() => {
            const san = status.parameters.find(p => p.key === 'SAN')
            const indefinite = san ? Math.floor((effectiveValue(san) * 4) / 5) : 0
            return (
              <>
                <div class="font-semibold text-slate-700 text-sm">不定</div>
                <div class="flex items-center justify-center h-8 font-semibold text-slate-800">{indefinite}</div>
                <div></div>
              </>
            )
          })()}
        </div>
      </div>
    </div>
  )
}

// Main sheet page
app.get('/cthulhu6', async (c) => {
  const store = c.get('store')
  const [state, gallery] = await Promise.all([
    buildSheetState(store),
    store.getImageGallery(CHAR_ID),
  ])

  return c.html(
    <Layout title="Character" headerActions={() => <SheetHeaderActions pc={state.pc} />}>
      <div class="grid grid-cols-1 max-w-[440px] mx-auto sm:grid-cols-[minmax(0,1fr)_320px] sm:max-w-none md:gap-2 md:max-w-3xl lg:grid-cols-[minmax(0,1fr)_648px] lg:gap-4 lg:max-w-[1104px] 2xl:grid-cols-[440px_968px] 2xl:max-w-[1536px]">
        <div class="flex flex-col gap-4">
          <Profile pc={state.pc} gallery={gallery} />
          <ScenarioMemoGroup pc={state.pc} />
        </div>
        <div class="flex flex-col gap-2">
          <div class="bg-white rounded-lg p-1 shadow">
            <input
              type="text"
              class="inline-block h-6 w-full text-xl font-semibold text-slate-800 placeholder:text-xl placeholder:font-semibold placeholder:text-slate-400 bg-transparent border-none outline-none"
              placeholder="クトゥルフ神話TRPG（第6版）"
            />
          </div>
          <div class="grid grid-cols-1 gap-2 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] 2xl:grid-cols-[minmax(0,1fr)_minmax(0,2fr)]">
            <StatusPanel state={state} />
            <SkillsPanel state={state} />
          </div>
        </div>
      </div>
      <PointsDisplay remaining={state.skills.remaining} />
      <GalleryModal pc={state.pc} gallery={gallery} />
    </Layout>
  )
})

// Header actions component (OOB version for preview toggle)
const SheetHeaderActionsOOB = ({ pc }: { pc: PageContext }) => (
  <div id="header-actions" class="flex items-center gap-2" hx-swap-oob="true">
    {pc.isOwner && (
      pc.preview ? (
        <Button
          variant="solid-blue"
          icon
          title="Exit preview mode"
          hxPost={api(pc, '/api/preview/off')}
          hxTarget="#memo-group"
          hxSwap="outerHTML"
        >
          <IconEye />
        </Button>
      ) : (
        <Button
          variant="ghost"
          icon
          title="Preview as visitor"
          hxPost={api(pc, '/api/preview/on')}
          hxTarget="#memo-group"
          hxSwap="outerHTML"
        >
          <IconEye />
        </Button>
      )
    )}
    <Button variant="ghost-blue" icon href="/characters" title="Character page">
      <IconAddressBook />
    </Button>
    <Button variant="solid-blue" icon title="Share options">
      <IconArrowUpFromBracket />
    </Button>
  </div>
)

// API: Preview mode toggle
app.post('/cthulhu6/api/preview/on', async (c) => {
  const store = c.get('store')
  const pc = await buildPageContext(store, true)
  return c.html(
    <>
      <MemoGroup pc={pc} />
      <ScenarioMemoGroup pc={pc} oob />
      <SheetHeaderActionsOOB pc={pc} />
    </>
  )
})

app.post('/cthulhu6/api/preview/off', async (c) => {
  const store = c.get('store')
  const pc = await buildPageContext(store, false)
  return c.html(
    <>
      <MemoGroup pc={pc} />
      <ScenarioMemoGroup pc={pc} oob />
      <SheetHeaderActionsOOB pc={pc} />
    </>
  )
})

// API: Set status variable
app.post('/cthulhu6/api/status/:key/set', async (c) => {
  const store = c.get('store')
  const rawKey = c.req.param('key')
  
  // Handle parameter set (param-HP, param-MP, param-SAN)
  if (rawKey.startsWith('param-')) {
    const paramKey = rawKey.replace('param-', '')
    const formData = await c.req.formData()
    const valueStr = formData.get(`param_${paramKey}`)
    if (!valueStr) return c.body(null, 204)
    
    const value = parseInt(String(valueStr), 10)
    if (isNaN(value)) return c.body(null, 204)
    
    await store.setParameter(CHAR_ID, paramKey, value)
    const state = await buildSheetState(store)
    return c.html(<StatusPanel state={state} oob />)
  }
  
  // Handle status variable set (status-STR, etc.)
  const key = rawKey.replace('status-', '')
  const formData = await c.req.formData()
  const valueStr = formData.get(`status_${key}`)
  
  if (!valueStr) {
    return c.body(null, 204)
  }

  const value = parseInt(String(valueStr), 10)
  if (isNaN(value)) {
    return c.body(null, 204)
  }

  const updated = await store.setVariableBase(CHAR_ID, key, value)
  if (!updated) {
    return c.body(null, 204)
  }

  const state = await buildSheetState(store)
  
  // DEX/EDU affect skill init values, INT affects skill points
  if (key === 'DEX' || key === 'EDU') {
    return c.html(
      <>
        <StatusPanel state={state} oob />
        <SkillsPanel state={state} oob />
        <PointsDisplay remaining={state.skills.remaining} oob />
      </>
    )
  }
  
  if (key === 'INT') {
    return c.html(
      <>
        <StatusPanel state={state} oob />
        <PointsDisplay remaining={state.skills.remaining} oob />
      </>
    )
  }

  return c.html(<StatusPanel state={state} oob />)
})

// API: Adjust status variable or parameter
app.post('/cthulhu6/api/status/:key/adjust', async (c) => {
  const store = c.get('store')
  const rawKey = c.req.param('key')
  const delta = parseInt(c.req.query('delta') || '0', 10)
  if (delta === 0) return c.body(null, 204)
  
  // Handle parameter adjust (param-HP, param-MP, param-SAN)
  if (rawKey.startsWith('param-')) {
    const paramKey = rawKey.replace('param-', '')
    await store.adjustParameter(CHAR_ID, paramKey, delta)
    const state = await buildSheetState(store)
    return c.html(<StatusPanel state={state} oob />)
  }
  
  // Handle status variable adjust (status-STR, etc.)
  const key = rawKey.replace('status-', '')
  const status = await store.getStatus(CHAR_ID)
  const variable = status.variables.find(v => v.key === key)
  if (!variable) return c.body(null, 204)
  
  const newValue = Math.max(variable.min, Math.min(variable.max, variable.base + delta))
  await store.setVariableBase(CHAR_ID, key, newValue)
  
  const state = await buildSheetState(store)
  
  // DEX/EDU affect skill init values
  if (key === 'DEX' || key === 'EDU') {
    return c.html(
      <>
        <StatusPanel state={state} oob />
        <SkillsPanel state={state} oob />
        <PointsDisplay remaining={state.skills.remaining} oob />
      </>
    )
  }
  
  // INT affects skill points
  if (key === 'INT') {
    return c.html(
      <>
        <StatusPanel state={state} oob />
        <PointsDisplay remaining={state.skills.remaining} oob />
      </>
    )
  }
  
  return c.html(<StatusPanel state={state} oob />)
})

// API: Set memo
app.post('/cthulhu6/api/memo/:id/set', async (c) => {
  const store = c.get('store')
  const id = c.req.param('id')
  const formData = await c.req.formData()
  const value = formData.get(id)
  
  await store.setMemo(CHAR_ID, id, String(value || ''))
  return c.body(null, 204)
})

// API: Set extra skill points
app.post('/cthulhu6/api/status/extra-job/set', async (c) => {
  const store = c.get('store')
  const formData = await c.req.formData()
  const valueStr = formData.get('extra_job')
  if (!valueStr) return c.body(null, 204)
  
  const value = parseInt(String(valueStr), 10)
  if (isNaN(value)) return c.body(null, 204)
  
  await store.setExtraJob(CHAR_ID, value)
  const state = await buildSheetState(store)
  return c.html(<ExtraPointsUpdateFragments state={state} />)
})

app.post('/cthulhu6/api/status/extra-hobby/set', async (c) => {
  const store = c.get('store')
  const formData = await c.req.formData()
  const valueStr = formData.get('extra_hobby')
  if (!valueStr) return c.body(null, 204)
  
  const value = parseInt(String(valueStr), 10)
  if (isNaN(value)) return c.body(null, 204)
  
  await store.setExtraHobby(CHAR_ID, value)
  const state = await buildSheetState(store)
  return c.html(<ExtraPointsUpdateFragments state={state} />)
})

// API: Adjust extra skill points
app.post('/cthulhu6/api/status/extra-job/adjust', async (c) => {
  const store = c.get('store')
  const delta = parseInt(c.req.query('delta') || '0', 10)
  if (delta === 0) return c.body(null, 204)
  
  await store.adjustExtraJob(CHAR_ID, delta)
  const state = await buildSheetState(store)
  return c.html(<ExtraPointsUpdateFragments state={state} />)
})

app.post('/cthulhu6/api/status/extra-hobby/adjust', async (c) => {
  const store = c.get('store')
  const delta = parseInt(c.req.query('delta') || '0', 10)
  if (delta === 0) return c.body(null, 204)
  
  await store.adjustExtraHobby(CHAR_ID, delta)
  const state = await buildSheetState(store)
  return c.html(<ExtraPointsUpdateFragments state={state} />)
})

// API: Skill grow toggle
app.post('/cthulhu6/api/skill/:key/grow', async (c) => {
  const store = c.get('store')
  const key = c.req.param('key')
  const skill = await store.getSkill(CHAR_ID, key)
  if (!skill || !skill.single) {
    return c.body(null, 204)
  }
  
  await store.updateSkill(CHAR_ID, key, { grow: !skill.single.grow })
  const state = await buildSheetState(store)
  return c.html(<SkillGrowUpdateFragments state={state} skillKey={key} />)
})

// API: Adjust skill value
app.post('/cthulhu6/api/skill/:key/:field/adjust', async (c) => {
  const store = c.get('store')
  const key = c.req.param('key')
  const field = c.req.param('field') as 'job' | 'hobby' | 'perm' | 'temp'
  const delta = parseInt(c.req.query('delta') || '0', 10)
  
  const skill = await store.getSkill(CHAR_ID, key)
  if (!skill || !skill.single) {
    return c.body(null, 204)
  }
  
  const newValue = Math.max(0, (skill.single[field] || 0) + delta)
  await store.updateSkill(CHAR_ID, key, { [field]: newValue })
  
  const state = await buildSheetState(store)
  return c.html(<SkillUpdateFragments state={state} skillKey={key} />)
})

// ========== Multi-Genre Skill API Routes ==========

// API: Add genre
app.post('/cthulhu6/api/skill/:key/genre/add', async (c) => {
  const store = c.get('store')
  const key = c.req.param('key')
  const result = await store.addGenre(CHAR_ID, key)
  if (!result) return c.body(null, 204)
  
  const state = await buildSheetState(store)
  return c.html(<SkillMultiPanelFragment state={state} skillKey={key} />)
})

// API: Delete genre
app.post('/cthulhu6/api/skill/:key/genre/:index/delete', async (c) => {
  const store = c.get('store')
  const key = c.req.param('key')
  const index = parseInt(c.req.param('index'), 10)
  if (isNaN(index)) return c.body(null, 204)
  
  const result = await store.deleteGenre(CHAR_ID, key, index)
  if (!result) return c.body(null, 204)
  
  const state = await buildSheetState(store)
  return c.html(<SkillMultiPanelFragment state={state} skillKey={key} />)
})

// API: Toggle genre grow
app.post('/cthulhu6/api/skill/:key/genre/:index/grow', async (c) => {
  const store = c.get('store')
  const key = c.req.param('key')
  const index = parseInt(c.req.param('index'), 10)
  if (isNaN(index)) return c.body(null, 204)
  
  const skill = await store.getSkill(CHAR_ID, key)
  if (!skill || !skill.multi) return c.body(null, 204)
  if (index < 0 || index >= skill.multi.genres.length) return c.body(null, 204)
  
  const currentGrow = skill.multi.genres[index].grow
  await store.updateGenreGrow(CHAR_ID, key, index, !currentGrow)
  
  const state = await buildSheetState(store)
  return c.html(<GenreGrowUpdateFragments state={state} skillKey={key} genreIndex={index} />)
})

// API: Set genre label
app.post('/cthulhu6/api/skill/:key/genre/:index/label', async (c) => {
  const store = c.get('store')
  const key = c.req.param('key')
  const index = parseInt(c.req.param('index'), 10)
  if (isNaN(index)) return c.body(null, 204)
  
  const formData = await c.req.formData()
  const label = formData.get(`genre_label_${key}_${index}`)
  if (label === null) return c.body(null, 204)
  
  await store.updateGenreLabel(CHAR_ID, key, index, String(label))
  return c.body(null, 204) // No OOB update needed for label
})

// API: Adjust genre field
app.post('/cthulhu6/api/skill/:key/genre/:index/:field/adjust', async (c) => {
  const store = c.get('store')
  const key = c.req.param('key')
  const index = parseInt(c.req.param('index'), 10)
  const field = c.req.param('field') as 'job' | 'hobby' | 'perm' | 'temp'
  const delta = parseInt(c.req.query('delta') || '0', 10)
  
  if (isNaN(index)) return c.body(null, 204)
  
  await store.updateGenreField(CHAR_ID, key, index, field, delta)
  
  const state = await buildSheetState(store)
  return c.html(<GenreUpdateFragments state={state} skillKey={key} genreIndex={index} />)
})

// ========== Custom Skill API Routes ==========

// API: Add custom skill
app.post('/cthulhu6/api/skill/custom/add', async (c) => {
  const store = c.get('store')
  await store.addCustomSkill(CHAR_ID)
  const state = await buildSheetState(store)
  return c.html(<CustomSkillsSectionFragment state={state} />)
})

// API: Delete custom skill
app.post('/cthulhu6/api/skill/custom/:index/delete', async (c) => {
  const store = c.get('store')
  const index = parseInt(c.req.param('index'), 10)
  if (isNaN(index)) return c.body(null, 204)
  
  const result = await store.deleteCustomSkill(CHAR_ID, index)
  if (!result) return c.body(null, 204)
  
  const state = await buildSheetState(store)
  return c.html(<CustomSkillsSectionFragment state={state} />)
})

// API: Toggle custom skill grow
app.post('/cthulhu6/api/skill/custom/:index/grow', async (c) => {
  const store = c.get('store')
  const index = parseInt(c.req.param('index'), 10)
  if (isNaN(index)) return c.body(null, 204)
  
  const customSkills = await store.getCustomSkills(CHAR_ID)
  if (index < 0 || index >= customSkills.length) return c.body(null, 204)
  
  const currentGrow = customSkills[index].grow
  await store.updateCustomSkillGrow(CHAR_ID, index, !currentGrow)
  
  const state = await buildSheetState(store)
  return c.html(<CustomSkillGrowUpdateFragments state={state} index={index} />)
})

// API: Set custom skill name
app.post('/cthulhu6/api/skill/custom/:index/name', async (c) => {
  const store = c.get('store')
  const index = parseInt(c.req.param('index'), 10)
  if (isNaN(index)) return c.body(null, 204)
  
  const formData = await c.req.formData()
  const name = formData.get(`custom_skill_name_${index}`)
  if (name === null) return c.body(null, 204)
  
  await store.updateCustomSkillName(CHAR_ID, index, String(name))
  return c.body(null, 204) // No OOB update needed for name
})

// API: Adjust custom skill field
app.post('/cthulhu6/api/skill/custom/:index/:field/adjust', async (c) => {
  const store = c.get('store')
  const index = parseInt(c.req.param('index'), 10)
  const field = c.req.param('field') as 'job' | 'hobby' | 'perm' | 'temp'
  const delta = parseInt(c.req.query('delta') || '0', 10)
  
  if (isNaN(index)) return c.body(null, 204)
  
  await store.updateCustomSkillField(CHAR_ID, index, field, delta)
  
  const state = await buildSheetState(store)
  return c.html(<CustomSkillUpdateFragments state={state} index={index} />)
})

// ========== Parameter API Routes ==========

// API: Set parameter value
app.post('/cthulhu6/api/param/:key/set', async (c) => {
  const store = c.get('store')
  const key = c.req.param('key').replace('param-', '')
  const formData = await c.req.formData()
  const valueStr = formData.get(`param_${key}`)
  
  if (!valueStr) return c.body(null, 204)
  const value = parseInt(String(valueStr), 10)
  if (isNaN(value)) return c.body(null, 204)
  
  await store.setParameter(CHAR_ID, key, value)
  const state = await buildSheetState(store)
  return c.html(<StatusPanel state={state} oob />)
})

// API: Adjust parameter value
app.post('/cthulhu6/api/param/:key/adjust', async (c) => {
  const store = c.get('store')
  const key = c.req.param('key').replace('param-', '')
  const delta = parseInt(c.req.query('delta') || '0', 10)
  if (delta === 0) return c.body(null, 204)
  
  await store.adjustParameter(CHAR_ID, key, delta)
  const state = await buildSheetState(store)
  return c.html(<StatusPanel state={state} oob />)
})

// ========== Profile API Routes ==========

// API: Set profile name
app.post('/cthulhu6/api/profile/name/set', async (c) => {
  const store = c.get('store')
  const formData = await c.req.formData()
  const name = formData.get('profile_name')
  if (name === null) return c.body(null, 204)
  
  await store.setProfileName(CHAR_ID, String(name))
  return c.body(null, 204)
})

// API: Set profile ruby
app.post('/cthulhu6/api/profile/ruby/set', async (c) => {
  const store = c.get('store')
  const formData = await c.req.formData()
  const ruby = formData.get('profile_ruby')
  if (ruby === null) return c.body(null, 204)
  
  await store.setProfileRuby(CHAR_ID, String(ruby))
  return c.body(null, 204)
})

// ========== Damage Bonus API Route ==========

// API: Set damage bonus
app.post('/cthulhu6/api/param/db/set', async (c) => {
  const store = c.get('store')
  const formData = await c.req.formData()
  const value = formData.get('param_db')
  if (value === null) return c.body(null, 204)
  
  await store.setDamageBonus(CHAR_ID, String(value))
  return c.body(null, 204)
})

// ========== Random Stats API Route ==========

// API: Randomize stats
app.post('/cthulhu6/api/status/random', async (c) => {
  const store = c.get('store')
  await store.randomizeStats(CHAR_ID)
  const state = await buildSheetState(store)
  
  // Random affects status panel, skills (DEX affects 回避, EDU affects 母国語), and points
  return c.html(
    <>
      <StatusPanel state={state} oob />
      <SkillsPanel state={state} oob />
      <PointsDisplay remaining={state.skills.remaining} oob />
    </>
  )
})

// ========== Image Gallery API Routes ==========

// Helper: Returns both ImageGallery and GalleryModal OOB fragments
const ImageGalleryOOBResponse = ({ pc, gallery }: { pc: PageContext; gallery: ImageGalleryState }) => (
  <>
    <ImageGallery pc={pc} gallery={gallery} oob />
    <GalleryModal pc={pc} gallery={gallery} oob />
  </>
)

// API: Upload image
app.post('/cthulhu6/api/image/upload', async (c) => {
  const store = c.get('store')
  const formData = await c.req.formData()
  const file = formData.get('image') as File | null
  
  if (!file || typeof file === 'string') {
    return c.body(null, 400)
  }
  
  // Convert to base64 data URL for in-memory storage
  const arrayBuffer = await file.arrayBuffer()
  const base64 = btoa(String.fromCharCode(...new Uint8Array(arrayBuffer)))
  const dataUrl = `data:${file.type};base64,${base64}`
  
  await store.addImage(CHAR_ID, file.name, dataUrl, file.type)
  
  const [pc, gallery] = await Promise.all([
    buildPageContext(store),
    store.getImageGallery(CHAR_ID),
  ])
  return c.html(<ImageGalleryOOBResponse pc={pc} gallery={gallery} />)
})

// API: Delete current image
app.post('/cthulhu6/api/image/delete', async (c) => {
  const store = c.get('store')
  await store.deleteCurrentImage(CHAR_ID)
  
  const [pc, gallery] = await Promise.all([
    buildPageContext(store),
    store.getImageGallery(CHAR_ID),
  ])
  return c.html(<ImageGalleryOOBResponse pc={pc} gallery={gallery} />)
})

// API: Pin/unpin current image
app.post('/cthulhu6/api/image/pin', async (c) => {
  const store = c.get('store')
  await store.pinCurrentImage(CHAR_ID)
  
  const [pc, gallery] = await Promise.all([
    buildPageContext(store),
    store.getImageGallery(CHAR_ID),
  ])
  return c.html(<ImageGalleryOOBResponse pc={pc} gallery={gallery} />)
})

// API: Navigate to previous image
app.post('/cthulhu6/api/image/prev', async (c) => {
  const store = c.get('store')
  await store.navigateImage(CHAR_ID, 'prev')
  
  const [pc, gallery] = await Promise.all([
    buildPageContext(store),
    store.getImageGallery(CHAR_ID),
  ])
  return c.html(<ImageGalleryOOBResponse pc={pc} gallery={gallery} />)
})

// API: Navigate to next image
app.post('/cthulhu6/api/image/next', async (c) => {
  const store = c.get('store')
  await store.navigateImage(CHAR_ID, 'next')
  
  const [pc, gallery] = await Promise.all([
    buildPageContext(store),
    store.getImageGallery(CHAR_ID),
  ])
  return c.html(<ImageGalleryOOBResponse pc={pc} gallery={gallery} />)
})

// Select specific image by index (for gallery modal)
app.post('/cthulhu6/api/image/select/:index', async (c) => {
  const store = c.get('store')
  const index = parseInt(c.req.param('index'), 10)
  if (!isNaN(index)) {
    await store.setCurrentImageIndex(CHAR_ID, index)
  }
  
  const [pc, gallery] = await Promise.all([
    buildPageContext(store),
    store.getImageGallery(CHAR_ID),
  ])
  return c.html(<ImageGalleryOOBResponse pc={pc} gallery={gallery} />)
})

// Redirect root to cthulhu6
app.get('/', (c) => c.redirect('/cthulhu6'))

export default app
