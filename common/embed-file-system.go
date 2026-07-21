package common

import (
	"embed"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-contrib/static"
)

// Credit: https://github.com/gin-contrib/static/issues/19

type embedFileSystem struct {
	http.FileSystem
}

func (e *embedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	if err != nil {
		return false
	}
	return true
}

func (e *embedFileSystem) Open(name string) (http.File, error) {
	if name == "/" {
		// This will make sure the index page goes to NoRouter handler,
		// which will use the replaced index bytes with analytic codes.
		return nil, os.ErrNotExist
	}
	return e.FileSystem.Open(name)
}

func EmbedFolder(fsEmbed embed.FS, targetPath string) static.ServeFileSystem {
	efs, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return &embedFileSystem{
		FileSystem: http.FS(efs),
	}
}

// themeAwareFileSystem delegates to the appropriate embedded FS based on
// the current theme (via GetTheme). This enables runtime theme switching
// without restarting the server. When the active theme has no registered
// filesystem (or none was provided), it falls back to defaultFS.
type themeAwareFileSystem struct {
	defaultFS   static.ServeFileSystem
	alternatives map[string]static.ServeFileSystem
}

func (t *themeAwareFileSystem) pick() static.ServeFileSystem {
	if t.alternatives != nil {
		if fs, ok := t.alternatives[GetTheme()]; ok {
			return fs
		}
	}
	return t.defaultFS
}

func (t *themeAwareFileSystem) Exists(prefix string, path string) bool {
	return t.pick().Exists(prefix, path)
}

func (t *themeAwareFileSystem) Open(name string) (http.File, error) {
	return t.pick().Open(name)
}

// NewThemeAwareFS returns a theme-aware filesystem that serves assets from
// defaultFS unless the active theme matches one of the registered alternatives.
func NewThemeAwareFS(defaultFS static.ServeFileSystem, alternatives map[string]static.ServeFileSystem) static.ServeFileSystem {
	return &themeAwareFileSystem{defaultFS: defaultFS, alternatives: alternatives}
}
