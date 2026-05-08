export interface ScanRequest {
  rootPath: string
  maxDepth: number
  excludeRepoNames: string[]
}

export interface RepoInfo {
  name: string
  path: string
}

export interface GitLogRequest {
  repos: RepoInfo[]
  since: string
  until: string
  author: string
  noMerges: boolean
}

export interface GitCommit {
  repoName: string
  repoPath: string
  hash: string
  authorName: string
  authorEmail: string
  commitTime: string
  message: string
  files: string[]
}

export interface RepoError {
  repoName: string
  repoPath: string
  message: string
}

export interface GitLogResponse {
  commits: GitCommit[]
  errors: RepoError[]
}
