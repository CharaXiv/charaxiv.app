import type { FC } from 'hono/jsx'
import { IconAngleLeft, IconAngleRight, IconAnglesLeft, IconAnglesRight } from './icons'

type NumberInputProps = {
  id: string
  name: string
  value: number
  min?: number
  max?: number
  placeholder?: string
  readonly?: boolean
  basePath: string
  hxPost?: string
  hxTarget?: string
  hxSwap?: string
}

export const NumberInput: FC<NumberInputProps> = ({
  id,
  name,
  value,
  min,
  max,
  placeholder,
  readonly = false,
  basePath,
  hxPost,
  hxTarget,
  hxSwap = 'outerHTML',
}) => {
  const adjustPath = hxPost 
    ? `${basePath}${hxPost}` 
    : `${basePath}/api/status/${id}/adjust`
  const setPath = `${basePath}/api/status/${id}/set`
  const targetSelector = hxTarget || `#${id}-wrapper`

  const atMin = min !== undefined && value <= min
  const atMax = max !== undefined && value >= max

  return (
    <div class="flex items-center h-8 w-full" id={`${id}-wrapper`} hx-sync="this:queue last">
      <button
        type="button"
        class="shrink-0 w-7 h-7 inline-flex items-center justify-center border-none bg-transparent text-slate-400 rounded-sm cursor-pointer transition-colors duration-150 hover:bg-slate-100 hover:text-slate-600 disabled:text-slate-200 disabled:cursor-not-allowed touch-manipulation"
        title="-5"
        disabled={readonly || atMin}
        hx-on:before-request="adjustNumberInput(this, -5)"
        hx-post={`${adjustPath}?delta=-5`}
        hx-target={targetSelector}
        hx-swap={hxSwap}
      >
        <span class="w-3 h-3"><IconAnglesLeft /></span>
      </button>
      <button
        type="button"
        class="shrink-0 w-7 h-7 inline-flex items-center justify-center border-none bg-transparent text-slate-400 rounded-sm cursor-pointer transition-colors duration-150 hover:bg-slate-100 hover:text-slate-600 disabled:text-slate-200 disabled:cursor-not-allowed touch-manipulation"
        title="-1"
        disabled={readonly || atMin}
        hx-on:before-request="adjustNumberInput(this, -1)"
        hx-post={`${adjustPath}?delta=-1`}
        hx-target={targetSelector}
        hx-swap={hxSwap}
      >
        <span class="w-3 h-3"><IconAngleLeft /></span>
      </button>
      <input
        type="number"
        class="flex-1 h-full min-w-0 max-w-full w-[4ch] px-1 border-none bg-transparent text-base font-semibold text-center outline-none touch-manipulation placeholder:font-semibold placeholder:text-slate-300"
        id={id}
        name={name}
        value={value}
        placeholder={placeholder}
        min={min}
        max={max}
        readonly={readonly}
        {...(!readonly ? {
          'hx-post': setPath,
          'hx-trigger': 'input changed delay:300ms',
          'hx-include': 'this',
        } : {})}
      />
      <button
        type="button"
        class="shrink-0 w-7 h-7 inline-flex items-center justify-center border-none bg-transparent text-slate-400 rounded-sm cursor-pointer transition-colors duration-150 hover:bg-slate-100 hover:text-slate-600 disabled:text-slate-200 disabled:cursor-not-allowed touch-manipulation"
        title="+1"
        disabled={readonly || atMax}
        hx-on:before-request="adjustNumberInput(this, 1)"
        hx-post={`${adjustPath}?delta=1`}
        hx-target={targetSelector}
        hx-swap={hxSwap}
      >
        <span class="w-3 h-3"><IconAngleRight /></span>
      </button>
      <button
        type="button"
        class="shrink-0 w-7 h-7 inline-flex items-center justify-center border-none bg-transparent text-slate-400 rounded-sm cursor-pointer transition-colors duration-150 hover:bg-slate-100 hover:text-slate-600 disabled:text-slate-200 disabled:cursor-not-allowed touch-manipulation"
        title="+5"
        disabled={readonly || atMax}
        hx-on:before-request="adjustNumberInput(this, 5)"
        hx-post={`${adjustPath}?delta=5`}
        hx-target={targetSelector}
        hx-swap={hxSwap}
      >
        <span class="w-3 h-3"><IconAnglesRight /></span>
      </button>
    </div>
  )
}
