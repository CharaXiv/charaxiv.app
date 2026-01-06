import type { FC, PropsWithChildren } from 'hono/jsx'
import { Header } from './Header'

type LayoutProps = PropsWithChildren<{
  title: string
  headerActions?: FC
}>

export const Layout: FC<LayoutProps> = ({ title, headerActions: HeaderActions, children }) => {
  return (
    <html lang="ja">
      <head>
        <meta charset="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>{title} - CharaXiv</title>
        <link rel="stylesheet" href="/styles.css" />
        <script src="https://unpkg.com/htmx.org@2.0.4" defer></script>
        <script src="https://unpkg.com/idiomorph@0.7.4/dist/idiomorph-ext.min.js" defer></script>
        <script src="/components.js" defer></script>
        {/* EasyMDE for markdown editing */}
        <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css" />
        <link rel="stylesheet" href="https://unpkg.com/easymde@2.18.0/dist/easymde.min.css" />
        <script src="https://unpkg.com/easymde@2.18.0/dist/easymde.min.js" async></script>
        <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js" async></script>
      </head>
      <body class="bg-slate-100 text-slate-800 min-h-screen font-sans" hx-ext="morph">
        <Header>
          {HeaderActions && <HeaderActions />}
        </Header>
        <main>
          {children}
        </main>
        {/* Mobile: padding at bottom for fixed header */}
        <div class="h-12 bg-white sm:hidden"></div>
      </body>
    </html>
  )
}
