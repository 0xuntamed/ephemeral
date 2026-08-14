import type { SVGProps } from 'react'

type IconProps = SVGProps<SVGSVGElement>

const iconProps: IconProps = {
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 1.8,
  strokeLinecap: 'round',
  strokeLinejoin: 'round',
  'aria-hidden': true,
}

export function PlusIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="M12 5v14M5 12h14" />
    </svg>
  )
}

export function SearchIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <circle cx="11" cy="11" r="7" />
      <path d="m16 16 4 4" />
    </svg>
  )
}

export function ChevronIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="m8 10 4 4 4-4" />
    </svg>
  )
}

export function RefreshIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="M20 6v5h-5M4 18v-5h5" />
      <path d="M6.1 9A7 7 0 0 1 18.5 6.5L20 11M4 13l1.5 4.5A7 7 0 0 0 18 15" />
    </svg>
  )
}

export function BranchIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <circle cx="6" cy="5" r="2" />
      <circle cx="18" cy="6" r="2" />
      <circle cx="6" cy="19" r="2" />
      <path d="M6 7v10M8 13h3a7 7 0 0 0 7-5" />
    </svg>
  )
}

export function ArrowRightIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="M5 12h14m-5-5 5 5-5 5" />
    </svg>
  )
}

export function ArrowUpRightIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="M7 17 17 7M8 7h9v9" />
    </svg>
  )
}

export function ClockIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <circle cx="12" cy="12" r="8" />
      <path d="M12 8v5l3 2" />
    </svg>
  )
}

export function TrashIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="M5 7h14M9 7V4h6v3m2 0-1 13H8L7 7m3 4v5m4-5v5" />
    </svg>
  )
}

export function RetryIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="M20 7v5h-5" />
      <path d="M18.5 8.5A7 7 0 1 0 19 16" />
    </svg>
  )
}

export function ShieldIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="m12 3 8 3v5c0 5-3.4 8.3-8 10-4.6-1.7-8-5-8-10V6l8-3Z" />
      <path d="m9 12 2 2 4-4" />
    </svg>
  )
}

export function AlertIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="M12 3 2.5 20h19L12 3Z" />
      <path d="M12 9v5m0 3h.01" />
    </svg>
  )
}

export function MetricsIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="M5 19V9m7 10V5m7 14v-7" />
    </svg>
  )
}

export function CloseIcon(props: IconProps) {
  return (
    <svg {...iconProps} {...props}>
      <path d="m6 6 12 12M18 6 6 18" />
    </svg>
  )
}
