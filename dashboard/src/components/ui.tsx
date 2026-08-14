import type { ButtonHTMLAttributes, ReactNode, SVGProps } from 'react'
import type { Status } from '../types'
import { statusLabel } from '../lib/environment'
import { buttonClass, cn, type ButtonIntent, type ButtonSize } from './design'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  intent?: ButtonIntent
  size?: ButtonSize
}

export function Button({
  className,
  intent = 'primary',
  size = 'md',
  type = 'button',
  ...props
}: ButtonProps) {
  return <button className={cn(buttonClass(intent, size), className)} type={type} {...props} />
}

export function Logo({ className, ...props }: SVGProps<SVGSVGElement>) {
  return (
    <svg
      className={cn('size-9', className)}
      viewBox="0 0 44 44"
      fill="none"
      aria-hidden="true"
      {...props}
    >
      <path
        d="M22 3 38.5 12.5v19L22 41 5.5 31.5v-19L22 3Z"
        fill="currentColor"
        stroke="currentColor"
        strokeWidth="2"
      />
      <path d="M14 17.2 22 12l8 5.2v9.6L22 32l-8-5.2v-9.6Z" fill="#18110d" />
      <path d="m18 18.5 4-2.5 4 2.5v7L22 28l-4-2.5v-7Z" fill="#ffc772" />
    </svg>
  )
}

const statusClasses: Record<Status, string> = {
  REQUESTED: 'border-amber-deep/45 bg-amber/15 text-amber-deep',
  PLANNING: 'border-amber-deep/45 bg-amber/15 text-amber-deep',
  PROVISIONING: 'border-amber-deep/45 bg-amber/15 text-amber-deep',
  HEALTH_CHECK: 'border-amber-deep/45 bg-amber/15 text-amber-deep',
  READY: 'border-amber-deep/55 bg-amber/20 text-amber-deep',
  DESTROYING: 'border-paper-ink-dim/35 bg-paper-ink/7 text-paper-ink-dim',
  DESTROYED: 'border-paper-ink-dim/35 bg-paper-ink/7 text-paper-ink-dim',
  FAILED: 'border-ember-700/45 bg-ember-100 text-ember-800',
  DESTROY_FAILED: 'border-ember-700/45 bg-ember-100 text-ember-800',
  ORPHANED: 'border-ember-700/45 bg-ember-100 text-ember-800',
  FAILED_CLEANED: 'border-paper-ink-dim/35 bg-paper-ink/7 text-paper-ink-dim',
}

const statusDotClasses: Record<Status, string> = {
  REQUESTED: 'bg-amber',
  PLANNING: 'bg-amber',
  PROVISIONING: 'animate-pulse-warm bg-amber-deep',
  HEALTH_CHECK: 'animate-pulse-warm bg-amber-deep',
  READY: 'bg-amber-deep',
  DESTROYING: 'animate-pulse-warm bg-paper-ink-dim',
  DESTROYED: 'bg-paper-ink-dim',
  FAILED: 'bg-ember-700',
  DESTROY_FAILED: 'bg-ember-700',
  ORPHANED: 'bg-ember-700',
  FAILED_CLEANED: 'bg-paper-ink-dim',
}

export function StatusBadge({ status }: { status: Status }) {
  return (
    <span
      className={cn(
        'inline-flex min-h-7 items-center gap-2 border px-2.5 font-label text-[0.68rem] font-semibold uppercase tracking-[0.14em]',
        statusClasses[status],
      )}
    >
      <span className={cn('size-1.5 shrink-0', statusDotClasses[status])} />
      {statusLabel(status)}
    </span>
  )
}

export function Spinner({ className }: { className?: string }) {
  return (
    <span
      className={cn(
        'size-4 animate-spin rounded-full border-2 border-current border-r-transparent',
        className,
      )}
      aria-hidden="true"
    />
  )
}

export function StatTicket({
  label,
  value,
  detail,
  accent = false,
}: {
  label: string
  value: string
  detail: string
  accent?: boolean
}) {
  return (
    <article className="relative min-h-32 border-r border-paper-ink/20 p-5 last:border-r-0 sm:p-6">
      <div className="flex items-center justify-between gap-3 font-label text-[0.68rem] font-semibold uppercase tracking-[0.2em] text-paper-ink-dim">
        <span>{label}</span>
        <span className={cn('h-1 w-8', accent ? 'bg-ember-600' : 'bg-amber-deep')} />
      </div>
      <strong
        className={cn(
          'mt-3 block font-display text-4xl leading-none tracking-wide sm:text-5xl',
          accent ? 'text-ember-700' : 'text-paper-ink',
        )}
      >
        {value}
      </strong>
      <p className="mt-2 text-sm text-paper-ink-dim">{detail}</p>
    </article>
  )
}

export function CheckerStrip({ className }: { className?: string }) {
  return <div className={cn('checker-strip h-3 w-full', className)} aria-hidden="true" />
}

export function FieldLabel({ htmlFor, children }: { htmlFor: string; children: ReactNode }) {
  return (
    <label
      className="mb-1.5 block font-label text-[0.7rem] font-semibold uppercase tracking-[0.18em] text-paper-ink-dim"
      htmlFor={htmlFor}
    >
      {children}
    </label>
  )
}
