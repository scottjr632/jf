import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowUpRight, Clock, Flame, Star } from 'lucide-react'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { usePullRequests, useRepos } from '@/hooks/useGithubData'

export const Route = createFileRoute('/github/')({
  component: GithubOverview,
})

function GithubOverview() {
  const repos = useRepos()
  const pullRequests = usePullRequests()

  const openCount = pullRequests.filter((pr) => pr.status !== 'merged').length
  const reviewNeeded = pullRequests.filter(
    (pr) => pr.reviewState === 'needs-review',
  ).length
  const recentlyTouched = pullRequests.filter((pr) => pr.updatedAt.includes('h'))
    .length

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,_#f8fafc_0%,_#eef2f5_40%,_#e9eef2_100%)] text-slate-900">
      <section className="relative overflow-hidden px-6 pt-16 pb-12">
        <div className="absolute -top-24 right-0 h-72 w-72 rounded-full bg-emerald-200/40 blur-3xl" />
        <div className="absolute -bottom-16 left-12 h-64 w-64 rounded-full bg-sky-200/40 blur-3xl" />
        <div className="relative mx-auto max-w-6xl">
          <div className="flex flex-col gap-6">
            <Badge variant="mint" className="w-fit">
              GitHub signal cockpit
            </Badge>
            <h1 className="text-4xl md:text-6xl font-semibold tracking-tight">
              Track repos, focus reviews, and clear the inbox fast.
            </h1>
            <p className="text-lg text-slate-600 max-w-2xl">
              Tailored to the repos you own, with smart filters that surface the
              pull requests that actually need your attention.
            </p>
            <div className="flex flex-wrap gap-3">
              <Link
                to="/github/inbox"
                className={cn(
                  buttonVariants({ size: 'lg' }),
                  'bg-emerald-500 text-white hover:bg-emerald-600',
                )}
              >
                Open PR inbox
                <ArrowUpRight className="w-4 h-4" />
              </Link>
              <Link
                to="/"
                className={cn(
                  buttonVariants({ variant: 'outline', size: 'lg' }),
                  'border-slate-300',
                )}
              >
                View overview
              </Link>
            </div>
          </div>
        </div>
      </section>

      <section className="px-6 pb-8">
        <div className="mx-auto max-w-6xl grid gap-4 md:grid-cols-3">
          <Card className="motion-safe:animate-[fade-up_0.6s_ease-out]">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base text-slate-500">
                <Flame className="h-4 w-4 text-amber-500" />
                Open PRs
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-semibold">{openCount}</div>
              <p className="text-sm text-slate-500">Across your tracked repos.</p>
            </CardContent>
          </Card>
          <Card className="motion-safe:animate-[fade-up_0.6s_ease-out] [animation-delay:120ms]">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base text-slate-500">
                <Clock className="h-4 w-4 text-sky-500" />
                Review needed
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-semibold">{reviewNeeded}</div>
              <p className="text-sm text-slate-500">Unreviewed or waiting on you.</p>
            </CardContent>
          </Card>
          <Card className="motion-safe:animate-[fade-up_0.6s_ease-out] [animation-delay:240ms]">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base text-slate-500">
                <Star className="h-4 w-4 text-emerald-500" />
                Hot updates
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-semibold">{recentlyTouched}</div>
              <p className="text-sm text-slate-500">Updated in the last few hours.</p>
            </CardContent>
          </Card>
        </div>
      </section>

      <section className="px-6 pb-16">
        <div className="mx-auto max-w-6xl grid gap-6 lg:grid-cols-[1.5fr_1fr]">
          <Card className="border-slate-200/60">
            <CardHeader>
              <CardTitle>Repo pulse</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {repos.map((repo) => (
                <div
                  key={repo.id}
                  className="flex flex-col gap-3 rounded-2xl border border-slate-200/60 bg-white/80 p-4 shadow-sm transition hover:-translate-y-0.5"
                >
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div>
                      <div className="flex items-center gap-2 text-lg font-semibold">
                        {repo.name}
                        <Badge variant={repo.visibility === 'public' ? 'mint' : 'slate'}>
                          {repo.visibility}
                        </Badge>
                      </div>
                      <p className="text-sm text-slate-500">{repo.description}</p>
                    </div>
                    <div className="text-sm text-slate-500">Updated {repo.updatedAt}</div>
                  </div>
                  <div className="flex flex-wrap items-center gap-2 text-sm text-slate-600">
                    <span className="rounded-full bg-slate-100 px-3 py-1">
                      {repo.language}
                    </span>
                    <span>★ {repo.stars}</span>
                    <span>Forks {repo.forks}</span>
                    <span>{repo.openPullRequests} open PRs</span>
                    <span>{repo.reviewVelocity}d avg review</span>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {repo.topics.map((topic) => (
                      <Badge key={topic} variant="sky">
                        {topic}
                      </Badge>
                    ))}
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>

          <Card className="border-slate-200/60">
            <CardHeader>
              <CardTitle>Priority review queue</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {pullRequests
                .filter((pr) => pr.reviewState === 'needs-review')
                .slice(0, 4)
                .map((pr) => (
                  <Link
                    key={pr.id}
                    to="/github/review/$prId"
                    params={{ prId: String(pr.id) }}
                    className="group flex flex-col gap-2 rounded-2xl border border-slate-200/60 bg-white/80 p-4 transition hover:-translate-y-0.5"
                  >
                    <div className="flex items-center justify-between text-sm text-slate-500">
                      <span>{pr.repo} · #{pr.number}</span>
                      <span>{pr.updatedAt}</span>
                    </div>
                    <div className="text-base font-semibold text-slate-900 group-hover:text-emerald-700">
                      {pr.title}
                    </div>
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span>{pr.additions} add</span>
                      <span>{pr.deletions} del</span>
                      <span>{pr.comments} comments</span>
                    </div>
                  </Link>
                ))}
              <Link
                to="/github/inbox"
                className={cn(
                  buttonVariants({ variant: 'outline', size: 'sm' }),
                  'w-full justify-center border-slate-200',
                )}
              >
                Open full inbox
              </Link>
            </CardContent>
          </Card>
        </div>
      </section>
    </div>
  )
}
