import type { FC, PropsWithChildren } from 'hono/jsx'

type ButtonVariant = 'ghost' | 'ghost-blue' | 'ghost-red' | 'ghost-green' | 'ghost-orange' | 'solid-blue' | 'solid-plain' | 'solid-orange' | 'outline-blue'

type ButtonProps = PropsWithChildren<{
  variant?: ButtonVariant
  icon?: boolean
  title?: string
  href?: string
  type?: 'button' | 'submit'
  disabled?: boolean
  class?: string
  as?: 'button' | 'span' // Allows rendering as span for use in labels
  // HTMX attributes
  hxPost?: string
  hxSwap?: string
  hxTarget?: string
  hxConfirm?: string
}>

const variantClasses: Record<ButtonVariant, string> = {
  'ghost': 'bg-transparent text-slate-600 hover:bg-slate-100 hover:text-slate-800',
  'ghost-blue': 'bg-transparent text-blue-600 hover:bg-blue-50 hover:text-blue-700',
  'ghost-red': 'bg-transparent text-red-600 hover:bg-red-50 hover:text-red-700',
  'ghost-green': 'bg-transparent text-green-600 hover:bg-green-50 hover:text-green-700',
  'ghost-orange': 'bg-transparent text-orange-600 hover:bg-orange-50 hover:text-orange-700',
  'solid-blue': 'bg-blue-600 text-white hover:bg-blue-700',
  'solid-plain': 'bg-slate-200 text-slate-600 hover:bg-slate-300',
  'solid-orange': 'bg-orange-500 text-white hover:bg-orange-600',
  'outline-blue': 'bg-transparent border border-blue-500 text-blue-600 hover:bg-blue-50 hover:text-blue-700',
}

export const Button: FC<ButtonProps> = ({ 
  variant = 'ghost', 
  icon = false, 
  title, 
  href, 
  type = 'button',
  disabled,
  class: className,
  as,
  hxPost,
  hxSwap,
  hxTarget,
  hxConfirm,
  children 
}) => {
  const baseClasses = 'inline-flex items-center justify-center rounded-md transition-colors duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed'
  const sizeClasses = icon ? 'w-9 h-9' : 'px-4 h-9 gap-2'
  const classes = `${baseClasses} ${sizeClasses} ${variantClasses[variant]} ${className || ''}`
  
  const htmxProps: Record<string, string> = {}
  if (hxPost) htmxProps['hx-post'] = hxPost
  if (hxSwap) htmxProps['hx-swap'] = hxSwap
  if (hxTarget) htmxProps['hx-target'] = hxTarget
  if (hxConfirm) htmxProps['hx-confirm'] = hxConfirm

  if (href) {
    return (
      <a href={href} class={classes} title={title}>
        {children}
      </a>
    )
  }

  // Render as span for use in labels (e.g., file upload buttons)
  if (as === 'span') {
    return (
      <span class={classes} title={title}>
        {children}
      </span>
    )
  }

  return (
    <button 
      type={type} 
      class={classes} 
      title={title}
      disabled={disabled}
      {...htmxProps}
    >
      {children}
    </button>
  )
}
