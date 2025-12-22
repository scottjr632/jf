import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center rounded-full border px-3 py-1 text-xs font-semibold uppercase tracking-wide transition',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-foreground/10 text-foreground',
        mint: 'border-emerald-500/30 bg-emerald-500/15 text-emerald-700',
        amber: 'border-amber-500/30 bg-amber-500/15 text-amber-700',
        sky: 'border-sky-500/30 bg-sky-500/15 text-sky-700',
        slate: 'border-slate-500/30 bg-slate-500/15 text-slate-700',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  )
}

export { Badge, badgeVariants }
