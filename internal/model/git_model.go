package model

type ScanRequest struct {
	RootPath         string   `json:"rootPath"`
	MaxDepth         int      `json:"maxDepth"`
	ExcludeRepoNames []string `json:"excludeRepoNames"`
}

type RepoInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type GitLogRequest struct {
	Repos    []RepoInfo `json:"repos"`
	Since    string     `json:"since"`
	Until    string     `json:"until"`
	Author   string     `json:"author"`
	NoMerges bool       `json:"noMerges"`
}

type GitCommit struct {
	RepoName    string   `json:"repoName"`
	RepoPath    string   `json:"repoPath"`
	Hash        string   `json:"hash"`
	AuthorName  string   `json:"authorName"`
	AuthorEmail string   `json:"authorEmail"`
	CommitTime  string   `json:"commitTime"`
	Message     string   `json:"message"`
	Files       []string `json:"files"`
}

type RepoError struct {
	RepoName string `json:"repoName"`
	RepoPath string `json:"repoPath"`
	Message  string `json:"message"`
}

type GitLogResponse struct {
	Commits []GitCommit `json:"commits"`
	Errors  []RepoError `json:"errors"`
}
