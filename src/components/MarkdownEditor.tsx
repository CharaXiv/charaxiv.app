import type { FC } from 'hono/jsx'
import type { PageContext } from '../lib/types'
import { isReadOnly, getMemo, api } from '../lib/types'

type MarkdownEditorEditProps = {
  id: string
  name: string
  placeholder: string
  value: string
  secret?: boolean
  basePath: string
}

// Just the edit textarea (used in memo groups)
export const MarkdownEditorEdit: FC<MarkdownEditorEditProps> = ({
  id,
  name,
  placeholder,
  value,
  secret = false,
  basePath,
}) => {
  return (
    <div class="bg-white rounded-lg overflow-hidden shadow" id={`${id}-wrapper`}>
      <textarea
        id={id}
        name={name}
        class="markdown-editor w-full min-h-[150px] p-3 border border-slate-200 rounded-md text-base leading-relaxed resize-y outline-none transition-colors duration-150 focus:border-blue-500"
        placeholder={placeholder}
        data-secret={secret || undefined}
        hx-post={`${basePath}/api/memo/${id}/set`}
        hx-trigger="input changed delay:3000ms"
        hx-swap="none"
      >{value}</textarea>
    </div>
  )
}

type MemoGroupProps = {
  pc: PageContext
  oob?: boolean
}

// Character-level memos (公開メモ, 秘匿メモ)
export const MemoGroup: FC<MemoGroupProps> = ({ pc, oob }) => {
  const readonly = isReadOnly(pc)
  
  // In readonly mode, show the rendered markdown
  if (readonly) {
    return (
      <div 
        id="memo-group" 
        class="flex flex-col gap-4"
        {...(oob ? { 'hx-swap-oob': 'true' } : {})}
      >
        <MarkdownRenderer title="公開メモ" content={getMemo(pc, 'public-memo')} />
        <MarkdownRenderer title="秘匿メモ" content={getMemo(pc, 'secret-memo')} secret />
      </div>
    )
  }
  
  return (
    <div 
      id="memo-group" 
      class="flex flex-col gap-4"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      <MarkdownEditorEdit
        id="public-memo"
        name="public-memo"
        placeholder="公開メモを入力..."
        value={getMemo(pc, 'public-memo')}
        basePath={pc.basePath}
      />
      <MarkdownEditorEdit
        id="secret-memo"
        name="secret-memo"
        placeholder="秘匿メモを入力..."
        value={getMemo(pc, 'secret-memo')}
        secret
        basePath={pc.basePath}
      />
    </div>
  )
}

// Scenario-level memos (シナリオ公開メモ, シナリオ秘匿メモ)
export const ScenarioMemoGroup: FC<MemoGroupProps> = ({ pc, oob }) => {
  const readonly = isReadOnly(pc)
  
  if (readonly) {
    return (
      <div 
        id="scenario-memo-group" 
        class="flex flex-col gap-4"
        {...(oob ? { 'hx-swap-oob': 'true' } : {})}
      >
        <MarkdownRenderer title="シナリオ公開メモ" content={getMemo(pc, 'scenario-public-memo')} />
        <MarkdownRenderer title="シナリオ秘匿メモ" content={getMemo(pc, 'scenario-secret-memo')} secret />
      </div>
    )
  }
  
  return (
    <div 
      id="scenario-memo-group" 
      class="flex flex-col gap-4"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      <MarkdownEditorEdit
        id="scenario-public-memo"
        name="scenario-public-memo"
        placeholder="シナリオ公開メモを入力..."
        value={getMemo(pc, 'scenario-public-memo')}
        basePath={pc.basePath}
      />
      <MarkdownEditorEdit
        id="scenario-secret-memo"
        name="scenario-secret-memo"
        placeholder="シナリオ秘匿メモを入力..."
        value={getMemo(pc, 'scenario-secret-memo')}
        secret
        basePath={pc.basePath}
      />
    </div>
  )
}

type MarkdownRendererProps = {
  title: string
  content: string
  secret?: boolean
}

// Rendered markdown for readonly mode
const MarkdownRenderer: FC<MarkdownRendererProps> = ({ title, content, secret = false }) => {
  return (
    <div class="markdown-renderer bg-white rounded-lg overflow-hidden shadow">
      <div class="flex items-center justify-between px-4 py-2 bg-slate-50 border-b border-slate-200">
        <div class="text-xl font-medium text-slate-500">{title}</div>
        {secret && (
          <button
            type="button"
            class="secret-toggle-btn inline-flex items-center justify-center w-8 h-8 border-none bg-transparent text-slate-500 rounded-md cursor-pointer transition-colors duration-150 hover:bg-slate-200 hover:text-slate-700"
            title="秘匿メモを表示"
          >
            {/* Eye icon (indicates content is hidden, click to show) */}
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 576 512" fill="currentColor" class="w-4 h-4">
              <path d="M288 32c-80.8 0-145.5 36.8-192.6 80.6C48.6 156 17.3 208 2.5 243.7c-3.3 7.9-3.3 16.7 0 24.6C17.3 304 48.6 356 95.4 399.4C142.5 443.2 207.2 480 288 480s145.5-36.8 192.6-80.6c46.8-43.5 78.1-95.4 93-131.1c3.3-7.9 3.3-16.7 0-24.6c-14.9-35.7-46.2-87.7-93-131.1C433.5 68.8 368.8 32 288 32zM144 256a144 144 0 1 1 288 0 144 144 0 1 1 -288 0zm144-64c0 35.3-28.7 64-64 64c-7.1 0-13.9-1.2-20.3-3.3c-5.5-1.8-11.9 1.6-11.7 7.4c.3 6.9 1.3 13.8 3.2 20.7c13.7 51.2 66.4 81.6 117.6 67.9s81.6-66.4 67.9-117.6c-11.1-41.5-47.8-69.4-88.6-71.1c-5.8-.2-9.2 6.1-7.4 11.7c2.1 6.4 3.3 13.2 3.3 20.3z" />
            </svg>
          </button>
        )}
      </div>
      <div 
        class={`markdown-content min-h-[100px] w-full p-3 text-base leading-relaxed ${secret ? 'blur-sm select-none pointer-events-none' : ''}`}
        data-markdown={content}
        data-rendered={!content ? 'true' : undefined}
      >
        {/* Content will be rendered by JS */}
      </div>
    </div>
  )
}
