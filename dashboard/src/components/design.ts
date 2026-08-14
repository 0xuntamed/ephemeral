export function cn(...classes: Array<string | false | null | undefined>) {
  return classes.filter(Boolean).join(' ')
}

export type ButtonIntent = 'primary' | 'outline' | 'paper' | 'danger' | 'ghost'
export type ButtonSize = 'sm' | 'md' | 'icon'

const intentClasses: Record<ButtonIntent, string> = {
  primary: 'border-amber bg-amber text-paper-ink hover:border-amber-bright hover:bg-amber-bright',
  outline:
    'border-amber/70 bg-transparent text-amber-bright hover:border-amber-bright hover:bg-amber/10',
  paper:
    'border-paper-ink/35 bg-transparent text-paper-ink hover:border-paper-ink hover:bg-paper-ink/7',
  danger:
    'border-ember-600 bg-transparent text-ember-700 hover:border-ember-700 hover:bg-ember-100',
  ghost:
    'border-transparent bg-transparent text-stage-ink-dim hover:border-stage-line hover:text-stage-ink',
}

const sizeClasses: Record<ButtonSize, string> = {
  sm: 'min-h-9 px-3 py-2 text-[0.72rem]',
  md: 'min-h-11 px-4 py-2.5 text-xs',
  icon: 'size-10 p-0',
}

export function buttonClass(intent: ButtonIntent = 'primary', size: ButtonSize = 'md') {
  return cn(
    'inline-flex shrink-0 items-center justify-center gap-2 border font-label font-semibold uppercase tracking-[0.16em] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-amber-bright focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
    intentClasses[intent],
    sizeClasses[size],
  )
}

export const fieldClass =
  'min-h-11 w-full border border-paper-ink/35 bg-paper/65 px-3 py-2.5 text-sm text-paper-ink outline-none placeholder:text-paper-ink-dim/60 focus:border-amber-deep focus:ring-2 focus:ring-amber-deep/20'
