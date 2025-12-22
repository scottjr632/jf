import {
  createCollection,
  localOnlyCollectionOptions,
} from '@tanstack/react-db'
import {
  githubInboxFilters,
  githubPullRequests,
  githubRepos,
  InboxFilterSchema,
  PullRequestSchema,
  RepoSchema,
  type InboxFilter,
  type PullRequest,
  type Repo,
} from '@/data/github'
import { z } from 'zod'

const MessageSchema = z.object({
  id: z.number(),
  text: z.string(),
  user: z.string(),
})

export type Message = z.infer<typeof MessageSchema>

export const messagesCollection = createCollection(
  localOnlyCollectionOptions({
    getKey: (message) => message.id,
    schema: MessageSchema,
  }),
)

export const reposCollection = createCollection(
  localOnlyCollectionOptions({
    getKey: (repo) => repo.id,
    schema: RepoSchema,
  }),
)

export const pullRequestsCollection = createCollection(
  localOnlyCollectionOptions({
    getKey: (pr) => pr.id,
    schema: PullRequestSchema,
  }),
)

export const inboxFiltersCollection = createCollection(
  localOnlyCollectionOptions({
    getKey: (filter) => filter.id,
    schema: InboxFilterSchema,
  }),
)

let seededGithub = false
let inboxFilterId = githubInboxFilters.length + 1

export const seedGithubCollections = () => {
  if (seededGithub) return
  seededGithub = true

  githubRepos.forEach((repo) => reposCollection.insert(repo))
  githubPullRequests.forEach((pr) => pullRequestsCollection.insert(pr))
  githubInboxFilters.forEach((filter) => inboxFiltersCollection.insert(filter))
}

export const addInboxFilter = (filter: Omit<InboxFilter, 'id'>) => {
  const created: InboxFilter = { id: inboxFilterId++, ...filter }
  inboxFiltersCollection.insert(created)
  return created
}

export type { Repo, PullRequest, InboxFilter }
