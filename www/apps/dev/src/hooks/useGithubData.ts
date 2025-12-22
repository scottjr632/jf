import { useLiveQuery } from '@tanstack/react-db'

import {
  inboxFiltersCollection,
  pullRequestsCollection,
  reposCollection,
  seedGithubCollections,
  type InboxFilter,
  type PullRequest,
  type Repo,
} from '@/db-collections'

export type InboxQueryToken = {
  type: 'repo' | 'label' | 'author' | 'status' | 'review' | 'text'
  value: string
}

const reviewMap: Record<string, PullRequest['reviewState']> = {
  needed: 'needs-review',
  'needs-review': 'needs-review',
  approved: 'approved',
  changes: 'changes-requested',
  'changes-requested': 'changes-requested',
}

const statusMap: Record<string, PullRequest['status']> = {
  open: 'open',
  draft: 'draft',
  merged: 'merged',
}

const parseInboxQuery = (query: string) => {
  const tokens: InboxQueryToken[] = []

  query
    .split(/\s+/)
    .map((token) => token.trim())
    .filter(Boolean)
    .forEach((token) => {
      const lower = token.toLowerCase()
      if (lower.startsWith('repo:')) {
        tokens.push({ type: 'repo', value: lower.replace('repo:', '') })
        return
      }
      if (lower.startsWith('label:')) {
        tokens.push({ type: 'label', value: lower.replace('label:', '') })
        return
      }
      if (lower.startsWith('author:')) {
        tokens.push({ type: 'author', value: lower.replace('author:', '') })
        return
      }
      if (lower.startsWith('review:')) {
        tokens.push({ type: 'review', value: lower.replace('review:', '') })
        return
      }
      if (lower.startsWith('is:')) {
        tokens.push({ type: 'status', value: lower.replace('is:', '') })
        return
      }
      tokens.push({ type: 'text', value: lower })
    })

  return tokens
}

export const filterPullRequests = (prs: PullRequest[], query: string) => {
  const tokens = parseInboxQuery(query)

  if (tokens.length === 0) return prs

  const grouped = tokens.reduce(
    (acc, token) => {
      acc[token.type].push(token)
      return acc
    },
    {
      repo: [] as InboxQueryToken[],
      label: [] as InboxQueryToken[],
      author: [] as InboxQueryToken[],
      status: [] as InboxQueryToken[],
      review: [] as InboxQueryToken[],
      text: [] as InboxQueryToken[],
    },
  )

  return prs.filter((pr) => {
    if (
      grouped.repo.length > 0 &&
      !grouped.repo.some((token) => pr.repo.toLowerCase().includes(token.value))
    ) {
      return false
    }

    if (
      grouped.label.length > 0 &&
      !grouped.label.some((token) =>
        pr.labels.some((label) => label.toLowerCase() === token.value),
      )
    ) {
      return false
    }

    if (
      grouped.author.length > 0 &&
      !grouped.author.some((token) =>
        pr.author.toLowerCase().includes(token.value),
      )
    ) {
      return false
    }

    if (
      grouped.review.length > 0 &&
      !grouped.review.some((token) => {
        const reviewState = reviewMap[token.value]
        return reviewState ? pr.reviewState === reviewState : false
      })
    ) {
      return false
    }

    if (
      grouped.status.length > 0 &&
      !grouped.status.some((token) => {
        const status = statusMap[token.value]
        return status ? pr.status === status : false
      })
    ) {
      return false
    }

    if (
      grouped.text.length > 0 &&
      !grouped.text.every((token) =>
        pr.title.toLowerCase().includes(token.value),
      )
    ) {
      return false
    }

    return true
  })
}

export const useRepos = () => {
  seedGithubCollections()

  const { data = [] } = useLiveQuery((query) =>
    query.from({ repo: reposCollection }).select(({ repo }) => ({ ...repo })),
  )

  return data as Repo[]
}

export const usePullRequests = () => {
  seedGithubCollections()

  const { data = [] } = useLiveQuery((query) =>
    query.from({ pr: pullRequestsCollection }).select(({ pr }) => ({ ...pr })),
  )

  return data as PullRequest[]
}

export const useInboxFilters = () => {
  seedGithubCollections()

  const { data = [] } = useLiveQuery((query) =>
    query
      .from({ inbox: inboxFiltersCollection })
      .select(({ inbox }) => ({ ...inbox })),
  )

  return data as InboxFilter[]
}
