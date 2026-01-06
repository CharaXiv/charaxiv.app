import type { FC, PropsWithChildren } from 'hono/jsx'
import { IconLayerGroup, IconSpinner, LogoCharaXiv } from './icons'

export const Header: FC<PropsWithChildren> = ({ children }) => {
  return (
    <header class="fixed bottom-0 left-0 right-0 z-50 flex h-12 w-full items-center justify-between px-2 bg-white/10 backdrop-blur-xl sm:sticky sm:top-0 sm:bottom-auto sm:left-auto sm:right-auto sm:transition-colors sm:duration-150" id="header">
      <a href="/" class="inline-flex items-center gap-2 no-underline text-slate-800 text-2xl font-medium leading-none hover:text-slate-900" id="header-logo">
        <span class="icon-logo w-6 h-6">
          <IconLayerGroup />
        </span>
        <IconSpinner />
        <LogoCharaXiv />
      </a>
      <div class="flex items-center gap-2">
        {children}
      </div>
    </header>
  )
}

type HeaderActionsProps = PropsWithChildren<{
  id: string
  oob?: boolean
}>

export const HeaderActions: FC<HeaderActionsProps> = ({ id, oob, children }) => {
  return (
    <div 
      id={id} 
      class="flex items-center gap-2"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      {children}
    </div>
  )
}
