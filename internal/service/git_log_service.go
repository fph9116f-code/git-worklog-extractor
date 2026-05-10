package service

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"git-worklog-extractor/internal/model"
)

type GitLogService struct{}

func NewGitLogService() *GitLogService {
	return &GitLogService{}
}

func (s *GitLogService) Query(req model.GitLogRequest) model.GitLogResponse {
	response := model.GitLogResponse{
		Commits: make([]model.GitCommit, 0),
		Errors:  make([]model.RepoError, 0),
	}

	for _, repo := range req.Repos {
		commits, err := queryRepoLogs(repo, req)
		if err != nil {
			response.Errors = append(response.Errors, model.RepoError{
				RepoName: repo.Name,
				RepoPath: repo.Path,
				Message:  err.Error(),
			})
			continue
		}
		response.Commits = append(response.Commits, commits...)
	}

	return response
}

func queryRepoLogs(repo model.RepoInfo, req model.GitLogRequest) ([]model.GitCommit, error) {
	args := []string{
		"-C", repo.Path,
		"-c", "core.quotepath=false",
		"log",
		"--date=iso-strict",
		"--pretty=format:%H%x1f%an%x1f%ae%x1f%aI%x1f%s%x1e",
		"--name-only",
	}

	if strings.TrimSpace(req.Since) != "" {
		args = append(args, "--since="+strings.TrimSpace(req.Since))
	}
	if strings.TrimSpace(req.Until) != "" {
		args = append(args, "--until="+strings.TrimSpace(req.Until))
	}
	if strings.TrimSpace(req.Author) != "" {
		args = append(args, "--author="+strings.TrimSpace(req.Author))
	}
	if req.NoMerges {
		args = append(args, "--no-merges")
	}

	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, errors.New(msg)
	}

	return parseGitLogOutput(repo, string(output)), nil
}

func parseGitLogOutput(repo model.RepoInfo, output string) []model.GitCommit {
	records := strings.Split(output, "\x1e")
	commits := make([]model.GitCommit, 0, len(records))

	for _, record := range records {
		record = strings.Trim(record, "\r\n")
		if record == "" {
			continue
		}

		lines := strings.Split(record, "\n")
		header := strings.TrimRight(lines[0], "\r")
		fields := strings.Split(header, "\x1f")
		if len(fields) < 5 {
			continue
		}

		files := make([]string, 0)
		for _, line := range lines[1:] {
			file := strings.TrimSpace(strings.TrimRight(line, "\r"))
			if file != "" {
				files = append(files, file)
			}
		}

		commits = append(commits, model.GitCommit{
			RepoName:    repo.Name,
			RepoPath:    repo.Path,
			Hash:        fields[0],
			AuthorName:  fields[1],
			AuthorEmail: fields[2],
			CommitTime:  fields[3],
			Message:     fields[4],
			Files:       files,
		})
	}

	return commits
}
