package main

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

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	projectDirName       = "LogikaProjects"
	uniqueNameIndexLimit = 1000
)

// App struct
type App struct {
	ctx      context.Context
	projects map[string]*project.Controller
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		projects: make(map[string]*project.Controller),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.OnFileDrop(a.ctx, func(x, y int, paths []string) {

	})
}

func (a *App) SelectLogFileDialog() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select log file or existing project directory",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Logs",
				Pattern:     "*.log",
			},
		},
	})
}

func (a *App) SelectProjectDirectoryDialog() (string, error) {
	pDir, err := a.projectsDir()
	if err != nil {
		return "", err
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Select project directory",
		CanCreateDirectories: true,
		DefaultDirectory:     pDir,
	})
}

func (a *App) Import(src string) (string, error) {
	p := project.NewController()
	projectDir, err := a.projectsDir()
	if err != nil {
		return "", err
	}
	destDir, err := a.ensureUniqueProjectPath(filepath.Join(projectDir, filepath.Base(src)))
	if err != nil {
		return "", err
	}
	if err := p.Import(src, destDir); err != nil {
		return "", err
	}
	return destDir, nil
}

func (a *App) ensureUniqueProjectPath(projectPath string) (string, error) {
	if _, err := os.Stat(projectPath); errors.Is(err, os.ErrNotExist) {
		return projectPath, nil
	}
	for i := 1; i < uniqueNameIndexLimit; i++ {
		newPath := fmt.Sprintf("%s_(%d)", projectPath, i)
		if _, err := os.Stat(newPath); errors.Is(err, os.ErrNotExist) {
			return newPath, nil
		}
	}
	return "", errors.New("too many projects with same name")
}

// projectDirName makes the name of the project directory in user home directory based on given log file name.
func (a *App) projectsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	pDir := filepath.Join(homeDir, projectDirName)
	if err := os.MkdirAll(pDir, 0755); err != nil {
		return "", err
	}
	return pDir, nil
}

func (a *App) Open(projectDir string) (string, error) {
	p := project.NewController()
	err := p.Open(projectDir)
	if err != nil {
		return "", err
	}
	id := uuid.NewString()
	a.projects[id] = p
	return id, nil
}

func (a *App) LuaOutput(projectId string) string {
	return a.projects[projectId].LuaOutput()
}

func (a *App) Reload(projectId string) error {
	return a.projects[projectId].Open(a.projects[projectId].Dir())
}

func (a *App) LoadSettings(projectId string) (string, error) {
	return a.projects[projectId].Settings()
}

func (a *App) SaveSettings(projectId, settings string) error {
	return a.projects[projectId].SaveSettings(settings)
}

func (a *App) Close(id string) error {
	p := a.projects[id]
	if err := p.Close(); err != nil {
		return err
	}
	delete(a.projects, id)
	return nil
}

type ProjectResponse struct {
	Id  string `json:"id"`
	Dir string `json:"dir"`
}

func (a *App) Projects() string {
	var result []ProjectResponse
	for id, p := range a.projects {
		result = append(result, ProjectResponse{
			Id:  id,
			Dir: p.Dir(),
		})
	}
	b, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	return string(b)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) Query(projectId, filters string, offset, limit int) (string, error) {
	p, found := a.projects[projectId]
	if !found {
		return "", fmt.Errorf("project not found")
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

func (a *App) CountLines(projectId, filters string) (int, error) {
	p, found := a.projects[projectId]
	if !found {
		return 0, fmt.Errorf("project not found")
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

func (a *App) RunScript(projectId, script string) (string, error) {
	p, found := a.projects[projectId]
	if !found {
		return "", fmt.Errorf("project not found")
	}
	return p.RunScript(script)
}

func (a *App) PreviewTemplate(projectId string) string {
	p, found := a.projects[projectId]
	if !found {
		return ""
	}
	return p.PreviewTemplate()
}

func (a *App) Filters(projectId string) (string, error) {
	p, found := a.projects[projectId]
	if !found {
		return "", nil
	}
	return p.FiltersJSON()
}

func (a *App) OpenProjectDirectory(projectId string) error {
	p, found := a.projects[projectId]
	if !found {
		return errors.New("project not found")
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

func (a *App) shutdown(ctx context.Context) {
	for _, p := range a.projects {
		p.Close()
	}
}
