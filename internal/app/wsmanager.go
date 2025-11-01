package app

import (
	"fmt"
	"logika/internal/project"
	"sync"
)

type WSManager struct {
	items map[string]*project.Workspace
	mu    sync.Mutex
}

func NewWSManager() *WSManager {
	return &WSManager{
		items: make(map[string]*project.Workspace),
	}
}

func (w *WSManager) Add(p *project.Workspace) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, exists := w.items[p.ID()]; exists {
		return fmt.Errorf("project already opened: %s", p.Dir())
	}
	w.items[p.ID()] = p
	return nil
}

func (w *WSManager) Get(id string) (*project.Workspace, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	ws, ok := w.items[id]
	if !ok {
		return nil, fmt.Errorf("workspace not found: %s", id)
	}
	return ws, nil
}

func (w *WSManager) List() (workspaces []*project.Workspace) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, ws := range w.items {
		workspaces = append(workspaces, ws)
	}
	return workspaces
}

func (w *WSManager) Remove(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.items, id)
}
