import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowRight, BellDot, GitPullRequest, Layers3 } from 'lucide-react'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { usePullRequests, useRepos } from '@/hooks/useGithubData'

export const Route = createFileRoute('/')({ component: App })

function App() {
  const repos = useRepos()
  const pullRequests = usePullRequests()

  const openPrs = pullRequests.filter((pr) => pr.status !== 'merged')
  const reviewNeeded = pullRequests.filter(
    (pr) => pr.reviewState === 'needs-review',
  )

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,_#f8fafc_0%,_#edf2f5_38%,_#e6ebf0_100%)] text-slate-900">
      <section className="relative px-6 pt-20 pb-12 overflow-hidden">
        <div className="absolute -top-28 left-8 h-80 w-80 rounded-full bg-amber-200/40 blur-3xl motion-safe:animate-[float-slow_10s_ease-in-out_infinite]" />
        <div className="absolute -bottom-24 right-6 h-72 w-72 rounded-full bg-emerald-200/40 blur-3xl" />
        <div className="mx-auto max-w-6xl relative">
          <div className="flex flex-col gap-6">
            <Badge variant="amber" className="w-fit">
              Personal PR command center
            </Badge>
            <h1 className="text-4xl md:text-6xl font-semibold tracking-tight">
              PR Atlas keeps your reviews clean, fast, and focused.
            </h1>
            <p className="text-lg text-slate-600 max-w-2xl">
              Browse your repos, build custom inbox filters, and jump into the
              diff view when it is time to review.
            </p>
            <div className="flex flex-wrap gap-3">
              <Link
                to="/github"
                className={cn(
                  buttonVariants({ size: 'lg' }),
                  'bg-slate-900 text-white hover:bg-slate-800',
                )}
              >
                Explore repos
                <ArrowRight className="h-4 w-4" />
              </Link>
              <Link
                to="/github/inbox"
                className={cn(
                  buttonVariants({ variant: 'outline', size: 'lg' }),
                  'border-slate-300',
                )}
              >
                Open PR inbox
              </Link>
            </div>
          </div>
        </div>
      </section>

      <section className="px-6 pb-12">
        <div className="mx-auto max-w-6xl grid gap-4 md:grid-cols-3">
          <Card className="motion-safe:animate-[fade-up_0.6s_ease-out]">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base text-slate-500">
                <Layers3 className="h-4 w-4 text-emerald-500" />
                Repos tracked
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-semibold">{repos.length}</div>
              <p className="text-sm text-slate-500">Owned and watched.</p>
            </CardContent>
          </Card>
          <Card className="motion-safe:animate-[fade-up_0.6s_ease-out] [animation-delay:120ms]">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base text-slate-500">
                <GitPullRequest className="h-4 w-4 text-amber-500" />
                Open pull requests
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-semibold">{openPrs.length}</div>
              <p className="text-sm text-slate-500">Across all repos.</p>
            </CardContent>
          </Card>
          <Card className="motion-safe:animate-[fade-up_0.6s_ease-out] [animation-delay:240ms]">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base text-slate-500">
                <BellDot className="h-4 w-4 text-sky-500" />
                Needs review
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-semibold">{reviewNeeded.length}</div>
              <p className="text-sm text-slate-500">Waiting on you.</p>
            </CardContent>
          </Card>
        </div>
      </section>

      <section className="px-6 pb-16">
        <div className="mx-auto max-w-6xl grid gap-6 lg:grid-cols-[1.2fr_1fr]">
          <Card className="border-slate-200/60">
            <CardHeader>
              <CardTitle>Recent repos</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {repos.slice(0, 4).map((repo) => (
                <div
                  key={repo.id}
                  className="flex flex-col gap-3 rounded-2xl border border-slate-200/60 bg-white/80 p-4 shadow-sm transition hover:-translate-y-0.5"
                >
                  <div className="flex items-center justify-between">
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
                  <div className="flex flex-wrap items-center gap-2 text-xs text-slate-600">
                    <span className="rounded-full bg-slate-100 px-3 py-1">
                      {repo.language}
                    </span>
                    <span>{repo.openPullRequests} open PRs</span>
                  </div>
                </div>
              ))}
              <Link
                to="/github"
                className={cn(
                  buttonVariants({ variant: 'outline', size: 'sm' }),
                  'w-full justify-center border-slate-200',
                )}
              >
                View all repos
              </Link>
            </CardContent>
          </Card>

          <Card className="border-slate-200/60">
            <CardHeader>
              <CardTitle>Inbox spotlight</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {reviewNeeded.slice(0, 4).map((pr) => (
                <Link
                  key={pr.id}
                  to="/github/review/$prId"
                  params={{ prId: String(pr.id) }}
                  className="group flex flex-col gap-2 rounded-2xl border border-slate-200/60 bg-white/90 p-4 transition hover:-translate-y-0.5"
                >
                  <div className="flex items-center justify-between text-sm text-slate-500">
                    <span>
                      {pr.repo} · #{pr.number}
                    </span>
                    <span>{pr.updatedAt}</span>
                  </div>
                  <div className="text-base font-semibold text-slate-900 group-hover:text-emerald-700">
                    {pr.title}
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {pr.labels.map((label) => (
                      <Badge key={label} variant="sky">
                        {label}
                      </Badge>
                    ))}
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
                Open inbox
              </Link>
            </CardContent>
          </Card>
        </div>
      </section>
    </div>
  )
}
