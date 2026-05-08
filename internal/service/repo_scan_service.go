package service

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"git-worklog-extractor2/internal/model"
)

type RepoScanService struct{}

func NewRepoScanService() *RepoScanService {
	return &RepoScanService{}
}

func (s *RepoScanService) Scan(req model.ScanRequest) ([]model.RepoInfo, error) {
	rootPath := strings.TrimSpace(req.RootPath)
	if rootPath == "" {
		return nil, os.ErrInvalid
	}

	rootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrInvalid
	}

	maxDepth := req.MaxDepth
	if maxDepth < 0 {
		maxDepth = 0
	}

	excludes := make(map[string]struct{}, len(req.ExcludeRepoNames))
	for _, name := range req.ExcludeRepoNames {
		name = strings.TrimSpace(name)
		if name != "" {
			excludes[strings.ToLower(name)] = struct{}{}
		}
	}

	repos := make([]model.RepoInfo, 0)
	err = scanRepoDir(rootPath, 0, maxDepth, excludes, &repos)
	if err != nil {
		return nil, err
	}

	sort.Slice(repos, func(i, j int) bool {
		return strings.ToLower(repos[i].Path) < strings.ToLower(repos[j].Path)
	})

	return repos, nil
}

func scanRepoDir(currentPath string, depth int, maxDepth int, excludes map[string]struct{}, repos *[]model.RepoInfo) error {
	name := filepath.Base(currentPath)
	if _, excluded := excludes[strings.ToLower(name)]; excluded {
		return nil
	}

	if isGitRepo(currentPath) {
		*repos = append(*repos, model.RepoInfo{
			Name: name,
			Path: currentPath,
		})
		return nil
	}

	if depth >= maxDepth {
		return nil
	}

	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Name() == ".git" {
			continue
		}
		if err := scanRepoDir(filepath.Join(currentPath, entry.Name()), depth+1, maxDepth, excludes, repos); err != nil {
			return err
		}
	}

	return nil
}

func isGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}
