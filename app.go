package main

import (
	"context"

	"git-worklog-extractor2/internal/model"
	"git-worklog-extractor2/internal/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	repoScanService *service.RepoScanService
	gitLogService   *service.GitLogService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		repoScanService: service.NewRepoScanService(),
		gitLogService:   service.NewGitLogService(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Git 仓库根目录",
	})
}

func (a *App) ScanRepositories(req model.ScanRequest) ([]model.RepoInfo, error) {
	return a.repoScanService.Scan(req)
}

func (a *App) QueryGitLogs(req model.GitLogRequest) (model.GitLogResponse, error) {
	return a.gitLogService.Query(req), nil
}
