import { createFileRoute, Link } from '@tanstack/react-router'
import { ChevronLeft, FileDiff, MessageSquareText } from 'lucide-react'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { usePullRequests } from '@/hooks/useGithubData'
import { pullRequestDiffs } from '@/data/github'

export const Route = createFileRoute('/github/review/$prId')({
  component: ReviewRoute,
})

function ReviewRoute() {
  const { prId } = Route.useParams()
  const pullRequests = usePullRequests()
  const pr = pullRequests.find((item) => String(item.id) === prId)
  const diffFiles = pr ? pullRequestDiffs[pr.id] ?? [] : []

  if (!pr) {
    return (
      <div className="min-h-screen bg-slate-950 text-white flex items-center justify-center">
        <div className="text-center space-y-4">
          <p className="text-xl">Pull request not found.</p>
          <Link
            to="/github/inbox"
            className={cn(buttonVariants({ variant: 'outline' }), 'border-white/30')}
          >
            Back to inbox
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-slate-950 text-white">
      <section className="px-6 pt-10 pb-6">
        <div className="mx-auto max-w-6xl space-y-6">
          <Link
            to="/github/inbox"
            className={cn(
              buttonVariants({ variant: 'ghost', size: 'sm' }),
              'text-slate-300 hover:bg-white/10',
            )}
          >
            <ChevronLeft className="h-4 w-4" />
            Back to inbox
          </Link>
          <div className="flex flex-col gap-4">
            <Badge
              variant="slate"
              className="w-fit border-white/20 bg-white/10 text-slate-200"
            >
              {pr.repo} · #{pr.number}
            </Badge>
            <h1 className="text-3xl md:text-4xl font-semibold">{pr.title}</h1>
            <div className="flex flex-wrap items-center gap-3 text-sm text-slate-300">
              <span>Author: {pr.author}</span>
              <span>Updated {pr.updatedAt}</span>
              <Badge
                variant="mint"
                className="border-emerald-400/40 bg-emerald-400/10 text-emerald-200"
              >
                {pr.reviewState.replace('-', ' ')}
              </Badge>
              <Badge
                variant={pr.status === 'merged' ? 'sky' : 'amber'}
                className={cn(
                  'border-white/20 bg-white/10 text-slate-200',
                  pr.status === 'merged'
                    ? 'border-sky-400/40 bg-sky-400/10 text-sky-200'
                    : 'border-amber-400/40 bg-amber-400/10 text-amber-200',
                )}
              >
                {pr.status}
              </Badge>
            </div>
          </div>
        </div>
      </section>

      <section className="px-6 pb-16">
        <div className="mx-auto max-w-6xl grid gap-6 lg:grid-cols-[2fr_1fr]">
          <div className="space-y-6">
            {diffFiles.length === 0 && (
              <Card className="border-white/10 bg-white/5 text-white">
                <CardHeader>
                  <CardTitle>No diff data yet</CardTitle>
                </CardHeader>
                <CardContent className="text-sm text-slate-300">
                  Wire this to GitHub's API to fetch real diffs for the PR.
                </CardContent>
              </Card>
            )}

            {diffFiles.map((file) => (
              <Card
                key={file.path}
                className="border-white/10 bg-white/5 text-white"
              >
                <CardHeader>
                  <CardTitle className="flex items-center justify-between text-sm">
                    <span className="flex items-center gap-2">
                      <FileDiff className="h-4 w-4 text-emerald-400" />
                      {file.path}
                    </span>
                    <span className="text-xs text-slate-400">
                      +{file.additions} / -{file.deletions}
                    </span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 font-mono text-sm">
                  {file.lines.map((line, index) => {
                    const style =
                      line.type === 'add'
                        ? 'bg-emerald-500/15 text-emerald-200'
                        : line.type === 'del'
                          ? 'bg-rose-500/20 text-rose-200'
                          : 'text-slate-300'

                    const prefix =
                      line.type === 'add' ? '+' : line.type === 'del' ? '-' : ' '

                    return (
                      <div
                        key={`${file.path}-${index}`}
                        className={cn('rounded-lg px-3 py-2', style)}
                      >
                        <span className="text-slate-400">{prefix}</span>{' '}
                        {line.value}
                      </div>
                    )
                  })}
                </CardContent>
              </Card>
            ))}
          </div>

          <Card className="border-white/10 bg-white/5 text-white h-fit">
            <CardHeader>
              <CardTitle>Review summary</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm text-slate-300">
              <div className="flex items-center justify-between">
                <span>Additions</span>
                <span className="text-white">{pr.additions}</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Deletions</span>
                <span className="text-white">{pr.deletions}</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Comments</span>
                <span className="text-white">{pr.comments}</span>
              </div>
              <div className="flex items-center justify-between">
                <span>Reviewers</span>
                <span className="text-white">{pr.reviewers.join(', ')}</span>
              </div>
              <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                <div className="flex items-center gap-2 text-xs uppercase text-slate-400">
                  <MessageSquareText className="h-4 w-4" />
                  Suggested next action
                </div>
                <p className="mt-2 text-sm text-slate-200">
                  Leave a high-level summary comment and approve once tests pass.
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </section>
    </div>
  )
}
