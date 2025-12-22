import { createFileRoute, Link } from '@tanstack/react-router'
import { useEffect, useMemo, useState } from 'react'
import {
  Filter,
  GitPullRequest,
  Sparkle,
  SquarePen,
  TimerReset,
} from 'lucide-react'

import { Badge } from '@/components/ui/badge'
import { Button, buttonVariants } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import {
  filterPullRequests,
  useInboxFilters,
  usePullRequests,
} from '@/hooks/useGithubData'
import { addInboxFilter } from '@/db-collections'

export const Route = createFileRoute('/github/inbox')({
  component: InboxRoute,
})

const accents = ['mint', 'amber', 'sky', 'slate'] as const

type Accent = (typeof accents)[number]

function InboxRoute() {
  const filters = useInboxFilters()
  const pullRequests = usePullRequests()

  const [selectedFilterId, setSelectedFilterId] = useState<number | null>(null)
  const [draftName, setDraftName] = useState('')
  const [draftQuery, setDraftQuery] = useState('label:priority review:needed')

  useEffect(() => {
    if (!selectedFilterId && filters.length > 0) {
      setSelectedFilterId(filters[0].id)
    }
  }, [filters, selectedFilterId])

  const activeFilter = filters.find((filter) => filter.id === selectedFilterId)
  const activeQuery = activeFilter?.query ?? ''

  const filteredPullRequests = useMemo(
    () => filterPullRequests(pullRequests, activeQuery),
    [pullRequests, activeQuery],
  )

  const previewPullRequests = useMemo(
    () => filterPullRequests(pullRequests, draftQuery),
    [pullRequests, draftQuery],
  )

  const handleCreateFilter = () => {
    if (!draftName.trim() || !draftQuery.trim()) return

    const accent =
      accents[filters.length % accents.length] ?? accents[0]

    const created = addInboxFilter({
      name: draftName.trim(),
      query: draftQuery.trim(),
      description: 'Custom inbox filter',
      accent: accent as Accent,
    })

    setDraftName('')
    setSelectedFilterId(created.id)
  }

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,_#f9fafb_0%,_#eef2f5_45%,_#e5eaee_100%)] text-slate-900">
      <section className="px-6 pt-14 pb-10">
        <div className="mx-auto max-w-6xl">
          <div className="flex flex-col gap-4">
            <Badge variant="sky" className="w-fit">
              PR inbox builder
            </Badge>
            <h1 className="text-4xl md:text-5xl font-semibold">
              Shape the inbox around what you want to review.
            </h1>
            <p className="text-lg text-slate-600 max-w-2xl">
              Save smart search filters and keep an always-on view of the pull
              requests you care about most.
            </p>
          </div>
        </div>
      </section>

      <section className="px-6 pb-16">
        <div className="mx-auto max-w-6xl grid gap-6 lg:grid-cols-[1fr_1.4fr]">
          <Card className="border-slate-200/60">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Filter className="h-4 w-4 text-emerald-500" />
                Saved filters
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-3">
                {filters.map((filter) => (
                  <button
                    key={filter.id}
                    onClick={() => setSelectedFilterId(filter.id)}
                    className={cn(
                      'w-full rounded-2xl border border-slate-200/60 bg-white/70 p-4 text-left transition hover:-translate-y-0.5',
                      selectedFilterId === filter.id
                        ? 'ring-2 ring-emerald-500/40'
                        : 'hover:ring-1 hover:ring-slate-200',
                    )}
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="text-base font-semibold">
                          {filter.name}
                        </div>
                        <p className="text-sm text-slate-500">
                          {filter.description}
                        </p>
                      </div>
                      <Badge variant={filter.accent}>{filter.accent}</Badge>
                    </div>
                    <div className="mt-3 rounded-xl bg-slate-100 px-3 py-2 text-xs font-mono text-slate-500">
                      {filter.query}
                    </div>
                  </button>
                ))}
              </div>

              <div className="rounded-2xl border border-dashed border-slate-300 bg-white/50 p-4">
                <div className="flex items-center gap-2 text-sm font-semibold text-slate-700">
                  <SquarePen className="h-4 w-4 text-emerald-500" />
                  Create a new filter
                </div>
                <div className="mt-3 space-y-3">
                  <Input
                    placeholder="Filter name"
                    value={draftName}
                    onChange={(event) => setDraftName(event.target.value)}
                  />
                  <Input
                    placeholder="Search query (label:priority review:needed)"
                    value={draftQuery}
                    onChange={(event) => setDraftQuery(event.target.value)}
                  />
                  <div className="text-xs text-slate-500">
                    Preview matches: {previewPullRequests.length} PRs
                  </div>
                  <Button
                    className="w-full"
                    onClick={handleCreateFilter}
                  >
                    Save filter
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-slate-200/60">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <GitPullRequest className="h-4 w-4 text-emerald-500" />
                {activeFilter ? activeFilter.name : 'Inbox'}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {activeFilter && (
                <div className="rounded-2xl border border-slate-200/70 bg-white/80 p-4">
                  <div className="flex items-center justify-between text-sm text-slate-500">
                    <span className="flex items-center gap-2">
                      <Sparkle className="h-4 w-4 text-amber-500" />
                      Active query
                    </span>
                    <span>{filteredPullRequests.length} matches</span>
                  </div>
                  <div className="mt-2 rounded-xl bg-slate-100 px-3 py-2 text-xs font-mono text-slate-500">
                    {activeQuery}
                  </div>
                </div>
              )}

              {filteredPullRequests.length === 0 && (
                <div className="rounded-2xl border border-dashed border-slate-300 bg-white/60 p-6 text-center text-sm text-slate-500">
                  No PRs match this filter yet.
                </div>
              )}

              {filteredPullRequests.map((pr) => (
                <Link
                  key={pr.id}
                  to="/github/review/$prId"
                  params={{ prId: String(pr.id) }}
                  className="group flex flex-col gap-3 rounded-2xl border border-slate-200/60 bg-white/90 p-4 transition hover:-translate-y-0.5"
                >
                  <div className="flex flex-wrap items-center justify-between gap-2 text-sm text-slate-500">
                    <span>
                      {pr.repo} · #{pr.number}
                    </span>
                    <span className="flex items-center gap-1">
                      <TimerReset className="h-4 w-4" />
                      {pr.updatedAt}
                    </span>
                  </div>
                  <div className="text-base font-semibold text-slate-900 group-hover:text-emerald-700">
                    {pr.title}
                  </div>
                  <div className="flex flex-wrap items-center gap-2 text-xs text-slate-500">
                    <Badge variant="mint">{pr.reviewState.replace('-', ' ')}</Badge>
                    <span>{pr.additions} add</span>
                    <span>{pr.deletions} del</span>
                    <span>{pr.comments} comments</span>
                  </div>
                </Link>
              ))}

              <Link
                to="/github"
                className={cn(
                  buttonVariants({ variant: 'outline', size: 'sm' }),
                  'w-full justify-center border-slate-200',
                )}
              >
                Back to repos
              </Link>
            </CardContent>
          </Card>
        </div>
      </section>
    </div>
  )
}
