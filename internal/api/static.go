package api

import (
	"net/http"
	"path/filepath"
	"strings"
)

type staticHandler struct {
	assetsDir http.Dir
}

func newStaticHandler(webDir string) *staticHandler {
	return &staticHandler{
		assetsDir: http.Dir(filepath.Join(webDir, "assets")),
	}
}

// FileSystem custom file system handler
type FileSystem struct {
	fs http.FileSystem
}

// Open opens file
func (fs FileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := fs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
}
