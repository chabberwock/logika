package userjs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type Manager struct {
	Dir string
}

func (m *Manager) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.RequestURI != "/dynamic/userjs.js" {
		return
	}
	scripts, err := m.combineScripts(m.Dir)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Header().Add("Content-Type", "application/javascript")
	io.Copy(writer, strings.NewReader(scripts))
}

func (m *Manager) combineScripts(dir string) (string, error) {
	var response string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if b, err := os.ReadFile(path.Join(dir, entry.Name())); err == nil {
			response += "\n" + string(b)
		} else {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
	}
	return response, nil
}
