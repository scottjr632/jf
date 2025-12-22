import { z } from 'zod'

export const RepoSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string(),
  visibility: z.enum(['public', 'private']),
  language: z.string(),
  stars: z.number(),
  forks: z.number(),
  updatedAt: z.string(),
  topics: z.array(z.string()),
  openPullRequests: z.number(),
  reviewVelocity: z.number(),
})

export type Repo = z.infer<typeof RepoSchema>

export const PullRequestSchema = z.object({
  id: z.number(),
  repo: z.string(),
  number: z.number(),
  title: z.string(),
  author: z.string(),
  status: z.enum(['open', 'draft', 'merged']),
  reviewState: z.enum(['needs-review', 'changes-requested', 'approved']),
  updatedAt: z.string(),
  additions: z.number(),
  deletions: z.number(),
  comments: z.number(),
  labels: z.array(z.string()),
  reviewers: z.array(z.string()),
})

export type PullRequest = z.infer<typeof PullRequestSchema>

export const InboxFilterSchema = z.object({
  id: z.number(),
  name: z.string(),
  query: z.string(),
  description: z.string(),
  accent: z.enum(['mint', 'amber', 'sky', 'slate']),
})

export type InboxFilter = z.infer<typeof InboxFilterSchema>

export type DiffLine = {
  type: 'add' | 'del' | 'context'
  value: string
}

export type DiffFile = {
  path: string
  additions: number
  deletions: number
  lines: DiffLine[]
}

export const githubRepos: Repo[] = [
  {
    id: 1,
    name: 'jf-web',
    description: 'Marketing site + docs with content automation.',
    visibility: 'public',
    language: 'TypeScript',
    stars: 182,
    forks: 28,
    updatedAt: '2h ago',
    topics: ['docs', 'design-system', 'mdx'],
    openPullRequests: 5,
    reviewVelocity: 1.6,
  },
  {
    id: 2,
    name: 'signal-ops',
    description: 'Internal tooling for release orchestration.',
    visibility: 'private',
    language: 'Go',
    stars: 0,
    forks: 3,
    updatedAt: 'Yesterday',
    topics: ['ops', 'automation', 'cli'],
    openPullRequests: 2,
    reviewVelocity: 2.4,
  },
  {
    id: 3,
    name: 'studio-ui',
    description: 'Component lab + shadcn based UI kit.',
    visibility: 'private',
    language: 'TSX',
    stars: 0,
    forks: 1,
    updatedAt: '3d ago',
    topics: ['design', 'storybook', 'ui'],
    openPullRequests: 3,
    reviewVelocity: 1.1,
  },
  {
    id: 4,
    name: 'infra-plans',
    description: 'Infrastructure templates and migrations.',
    visibility: 'public',
    language: 'HCL',
    stars: 64,
    forks: 9,
    updatedAt: '4d ago',
    topics: ['terraform', 'cloud', 'platform'],
    openPullRequests: 1,
    reviewVelocity: 3.2,
  },
  {
    id: 5,
    name: 'atlas-data',
    description: 'ETL pipeline + lakehouse experiments.',
    visibility: 'private',
    language: 'Python',
    stars: 0,
    forks: 2,
    updatedAt: '1w ago',
    topics: ['data', 'etl', 'analytics'],
    openPullRequests: 4,
    reviewVelocity: 4.7,
  },
  {
    id: 6,
    name: 'mobile-crest',
    description: 'Cross-platform client for the field team.',
    visibility: 'public',
    language: 'Kotlin',
    stars: 97,
    forks: 14,
    updatedAt: '2w ago',
    topics: ['mobile', 'sync', 'offline'],
    openPullRequests: 0,
    reviewVelocity: 0.8,
  },
]

export const githubPullRequests: PullRequest[] = [
  {
    id: 101,
    repo: 'jf-web',
    number: 482,
    title: 'Refine docs IA and add usage analytics',
    author: 'sara-m',
    status: 'open',
    reviewState: 'needs-review',
    updatedAt: '54m ago',
    additions: 214,
    deletions: 98,
    comments: 12,
    labels: ['docs', 'priority'],
    reviewers: ['you', 'alex-r'],
  },
  {
    id: 102,
    repo: 'signal-ops',
    number: 77,
    title: 'Release checklist automation + slack hooks',
    author: 'vito',
    status: 'open',
    reviewState: 'changes-requested',
    updatedAt: '3h ago',
    additions: 122,
    deletions: 34,
    comments: 6,
    labels: ['automation', 'ops'],
    reviewers: ['you'],
  },
  {
    id: 103,
    repo: 'studio-ui',
    number: 211,
    title: 'Add gradient button variants and motion tokens',
    author: 'lin',
    status: 'open',
    reviewState: 'needs-review',
    updatedAt: '6h ago',
    additions: 84,
    deletions: 12,
    comments: 2,
    labels: ['ui', 'design'],
    reviewers: ['you', 'maya'],
  },
  {
    id: 104,
    repo: 'jf-web',
    number: 479,
    title: 'Deprecate legacy pricing copy blocks',
    author: 'you',
    status: 'draft',
    reviewState: 'needs-review',
    updatedAt: '9h ago',
    additions: 55,
    deletions: 110,
    comments: 0,
    labels: ['content', 'cleanup'],
    reviewers: ['alex-r'],
  },
  {
    id: 105,
    repo: 'infra-plans',
    number: 19,
    title: 'Rotate S3 lifecycle configs for cold storage',
    author: 'ina',
    status: 'open',
    reviewState: 'approved',
    updatedAt: '1d ago',
    additions: 43,
    deletions: 18,
    comments: 4,
    labels: ['infra', 'platform'],
    reviewers: ['you', 'vito'],
  },
  {
    id: 106,
    repo: 'atlas-data',
    number: 301,
    title: 'Join strategy update for late arriving facts',
    author: 'marco',
    status: 'open',
    reviewState: 'needs-review',
    updatedAt: '2d ago',
    additions: 307,
    deletions: 144,
    comments: 9,
    labels: ['data', 'etl', 'priority'],
    reviewers: ['you', 'sara-m'],
  },
  {
    id: 107,
    repo: 'mobile-crest',
    number: 66,
    title: 'Offline queue reconciliation fixes',
    author: 'daria',
    status: 'merged',
    reviewState: 'approved',
    updatedAt: '4d ago',
    additions: 96,
    deletions: 61,
    comments: 3,
    labels: ['mobile', 'sync'],
    reviewers: ['you'],
  },
  {
    id: 108,
    repo: 'studio-ui',
    number: 205,
    title: 'Table density controls for dashboards',
    author: 'alex-r',
    status: 'open',
    reviewState: 'needs-review',
    updatedAt: '5d ago',
    additions: 148,
    deletions: 20,
    comments: 5,
    labels: ['ui', 'table'],
    reviewers: ['you'],
  },
]

export const githubInboxFilters: InboxFilter[] = [
  {
    id: 1,
    name: 'Priority review',
    query: 'label:priority review:needed',
    description: 'High-impact PRs that still need your review.',
    accent: 'amber',
  },
  {
    id: 2,
    name: 'Docs & UI',
    query: 'repo:jf-web repo:studio-ui label:docs label:ui',
    description: 'UI/Docs work across jf-web and studio-ui.',
    accent: 'sky',
  },
  {
    id: 3,
    name: 'Ops + Infra watchlist',
    query: 'label:ops label:infra is:open',
    description: 'Operational changes that require oversight.',
    accent: 'slate',
  },
]

export const pullRequestDiffs: Record<number, DiffFile[]> = {
  101: [
    {
      path: 'docs/overview.mdx',
      additions: 24,
      deletions: 6,
      lines: [
        { type: 'context', value: '## Why it exists' },
        { type: 'del', value: '- Instant onboarding for teams.' },
        { type: 'add', value: '+ Instant onboarding for cross-functional teams.' },
        { type: 'context', value: '' },
        { type: 'add', value: '### Usage analytics' },
        { type: 'add', value: '+ New dashboards show adoption in the first 30 days.' },
      ],
    },
    {
      path: 'apps/docs/analytics.ts',
      additions: 32,
      deletions: 8,
      lines: [
        { type: 'context', value: 'export const trackDocsView = (page: string) => {' },
        { type: 'del', value: '  analytics.track("Docs Viewed", { page })' },
        { type: 'add', value: '  analytics.track("Docs Viewed", { page, referrer: document.referrer })' },
        { type: 'add', value: '  analytics.track("Docs CTA", { page, placement: "footer" })' },
        { type: 'context', value: '}' },
      ],
    },
  ],
  103: [
    {
      path: 'packages/ui/button.tsx',
      additions: 18,
      deletions: 2,
      lines: [
        { type: 'context', value: 'const buttonVariants = cva(' },
        { type: 'add', value: '  { variant: { gradient: "bg-gradient-to-r from-emerald-500 to-teal-500" } }' },
        { type: 'context', value: '})' },
        { type: 'add', value: 'export const motionTokens = { hoverLift: "translate-y-[-2px]" }' },
      ],
    },
    {
      path: 'packages/ui/tokens.css',
      additions: 12,
      deletions: 0,
      lines: [
        { type: 'add', value: '.motion-lift { transition: transform 160ms ease; }' },
        { type: 'add', value: '.motion-lift:hover { transform: translateY(-2px); }' },
      ],
    },
  ],
  106: [
    {
      path: 'pipelines/facts/join.sql',
      additions: 42,
      deletions: 15,
      lines: [
        { type: 'context', value: 'WITH staged AS (' },
        { type: 'del', value: '  SELECT * FROM raw_facts WHERE loaded_at > now() - interval "7 days"' },
        { type: 'add', value: '  SELECT * FROM raw_facts WHERE loaded_at > now() - interval "14 days"' },
        { type: 'context', value: '),' },
        { type: 'add', value: 'late AS (SELECT * FROM raw_facts WHERE is_late = true),' },
        { type: 'add', value: 'final AS (SELECT * FROM staged UNION ALL SELECT * FROM late)' },
      ],
    },
  ],
}
