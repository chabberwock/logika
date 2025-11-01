package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"logika/internal/project"
	"os"
	"os/exec"
	"path/filepath"

	goRuntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	workspacesDirName    = "LogikaProjects"
	uniqueNameIndexLimit = 1000
)

// App struct
type App struct {
	ctx context.Context
	ws  *WSManager
}

// NewApp creates a new App application struct
func NewApp(ws *WSManager) *App {
	return &App{
		ws: ws,
	}
}

func (a *App) Start(ctx context.Context) {
	a.ctx = ctx
	runtime.OnFileDrop(a.ctx, func(x, y int, paths []string) {

	})
}

func (a *App) SelectLogFileDialog() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select log file or existing workspace directory",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Logs",
				Pattern:     "*.log",
			},
		},
	})
}

func (a *App) SelectWorkspaceDirectoryDialog() (string, error) {
	pDir, err := a.workspacesDir()
	if err != nil {
		return "", err
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Select workspace directory",
		CanCreateDirectories: true,
		DefaultDirectory:     pDir,
	})
}

func (a *App) Import(src string) (string, error) {
	p := project.NewController()
	workspaceDir, err := a.workspacesDir()
	if err != nil {
		return "", err
	}
	destDir, err := a.ensureUniqueWorkspacePath(filepath.Join(workspaceDir, filepath.Base(src)))
	if err != nil {
		return "", err
	}
	if err := p.Import(src, destDir); err != nil {
		return "", err
	}
	return destDir, nil
}

func (a *App) ensureUniqueWorkspacePath(workspacePath string) (string, error) {
	if _, err := os.Stat(workspacePath); errors.Is(err, os.ErrNotExist) {
		return workspacePath, nil
	}
	for i := 1; i < uniqueNameIndexLimit; i++ {
		newPath := fmt.Sprintf("%s_(%d)", workspacePath, i)
		if _, err := os.Stat(newPath); errors.Is(err, os.ErrNotExist) {
			return newPath, nil
		}
	}
	return "", errors.New("too many projects with same name")
}

// workspacesDirName makes the name of the project directory in user home directory based on given log file name.
func (a *App) workspacesDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	pDir := filepath.Join(homeDir, workspacesDirName)
	if err := os.MkdirAll(pDir, 0755); err != nil {
		return "", err
	}
	return pDir, nil
}

func (a *App) Open(workspaceDir string) (string, error) {
	p := project.NewController()
	err := p.Open(workspaceDir)
	if err != nil {
		return "", err
	}
	if err := a.ws.Add(p); err != nil {
		return "", err
	}
	return p.ID(), nil
}

func (a *App) LuaOutput(workspaceId string) string {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return err.Error()
	}
	return p.LuaOutput()
}

func (a *App) Reload(workspaceId string) error {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return err
	}
	return p.Open(p.Dir())
}

func (a *App) LoadSettings(workspaceId string) (string, error) {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return "", err
	}
	return p.Settings()
}

func (a *App) SaveSettings(workspaceId, settings string) error {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return err
	}
	return p.SaveSettings(settings)
}

func (a *App) Close(workspaceId string) error {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return err
	}
	a.ws.Remove(workspaceId)
	return p.Close()
}

type WorkspaceResponse struct {
	Id  string `json:"id"`
	Dir string `json:"dir"`
}

func (a *App) Workspaces() string {
	var result []WorkspaceResponse
	for _, p := range a.ws.List() {
		result = append(result, WorkspaceResponse{
			Id:  p.ID(),
			Dir: p.Dir(),
		})
	}
	b, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	return string(b)
}

func (a *App) Query(workspaceId, filters string, offset, limit int) (string, error) {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return "", err
	}
	var filtersMap map[string]any
	if err := json.Unmarshal([]byte(filters), &filtersMap); err != nil {
		return "", err
	}
	resp, err := p.Query(filtersMap, offset, limit)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (a *App) CountLines(workspaceId, filters string) (int, error) {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return 0, err
	}
	var filtersMap map[string]any
	if err := json.Unmarshal([]byte(filters), &filtersMap); err != nil {
		return 0, err
	}
	resp, err := p.Query(filtersMap, 0, 0)
	if err != nil {
		return 0, err
	}
	return len(resp), nil
}

func (a *App) RunScript(workspaceId, script string) (string, error) {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return "", err
	}
	return p.RunScript(script)
}

func (a *App) PreviewTemplate(workspaceId string) string {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return ""
	}
	return p.PreviewTemplate()
}

func (a *App) Filters(workspaceId string) (string, error) {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return "", err
	}
	return p.FiltersJSON()
}

func (a *App) OpenWorkspaceDirectory(workspaceId string) error {
	p, err := a.ws.Get(workspaceId)
	if err != nil {
		return err
	}
	var cmd *exec.Cmd

	switch goRuntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", p.Dir())
	case "darwin":
		cmd = exec.Command("open", p.Dir())
	case "linux":
		// Try xdg-open first (common), then fall back to gio open
		cmd = exec.Command("xdg-open", p.Dir())
	default:
		return fmt.Errorf("unsupported platform: %s", goRuntime.GOOS)
	}
	return cmd.Start()
}

func (a *App) Shutdown() {
	for _, p := range a.ws.List() {
		p.Close()
	}
}
