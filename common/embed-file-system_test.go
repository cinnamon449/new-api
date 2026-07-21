/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
package common

import (
	"net/http"
	"testing"

	"github.com/gin-contrib/static"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubFS is a minimal ServeFileSystem whose Exists/Open report a unique
// identity so tests can assert which FS the themeAwareFileSystem picked.
type stubFS struct {
	id  string
	ok  bool
	cfg stubFile
}

type stubFile struct{}

func (s *stubFS) Open(name string) (http.File, error) {
	return nil, nil
}

func (s *stubFS) Exists(prefix string, path string) bool {
	return s.ok
}

// newStubFS returns a stubFS that reports Exists=true so the dispatcher's
// return value can be observed via the embedded id on the returned object.
func newStubFS(id string) *stubFS {
	return &stubFS{id: id, ok: true, cfg: stubFile{}}
}

// withTheme sets the global active theme for the duration of a test and
// restores the previous value during cleanup. SetTheme is package-global,
// so tests using this helper must not call t.Parallel.
func withTheme(t *testing.T, theme string) {
	t.Helper()
	prev := GetTheme()
	SetTheme(theme)
	t.Cleanup(func() { themeValue.Store(prev) })
}

// TestSetThemeAllowList locks in the set of theme names the runtime
// switch accepts. Adding a new theme is a contract change that should
// be reflected here.
func TestSetThemeAllowList(t *testing.T) {
	withTheme(t, "default")
	require.Equal(t, "default", GetTheme())

	withTheme(t, "classic")
	require.Equal(t, "classic", GetTheme())

	withTheme(t, "interapi")
	require.Equal(t, "interapi", GetTheme())
}

// TestSetThemeRejectsUnknown ensures the runtime gate silently ignores
// typo'd or unregistered theme names instead of crashing on the empty
// FS branch — the value already stored must remain.
func TestSetThemeRejectsUnknown(t *testing.T) {
	withTheme(t, "default")
	SetTheme("bogus")
	assert.Equal(t, "default", GetTheme(), "unknown theme names must be ignored")
	SetTheme("")
	assert.Equal(t, "default", GetTheme(), "empty theme name must be ignored")
}

// TestThemeAwareFSPicksRegisteredAlternative verifies the dispatcher
// routes Exists/Open to the alternative registered for the active theme.
func TestThemeAwareFSPicksRegisteredAlternative(t *testing.T) {
	defaultFS := newStubFS("default")
	classicFS := newStubFS("classic")
	interapiFS := newStubFS("interapi")
	fs := NewThemeAwareFS(defaultFS, map[string]static.ServeFileSystem{
		"classic":  classicFS,
		"interapi": interapiFS,
	})
	ta, ok := fs.(*themeAwareFileSystem)
	require.True(t, ok, "NewThemeAwareFS must return *themeAwareFileSystem")

	withTheme(t, "interapi")
	require.Same(t, interapiFS, ta.pick(), "interapi theme must pick interapi FS")

	withTheme(t, "classic")
	require.Same(t, classicFS, ta.pick(), "classic theme must pick classic FS")
}

// TestThemeAwareFSFallsBackToDefault ensures that when the active theme
// has no registered FS (either "default" or an unrecognized name), the
// default FS is used. This protects the SPA from blank pages when an
// operator flips to a theme the binary wasn't built with.
func TestThemeAwareFSFallsBackToDefault(t *testing.T) {
	defaultFS := newStubFS("default")
	classicFS := newStubFS("classic")
	fs := NewThemeAwareFS(defaultFS, map[string]static.ServeFileSystem{
		"classic": classicFS,
	})
	ta, ok := fs.(*themeAwareFileSystem)
	require.True(t, ok)

	withTheme(t, "default")
	require.Same(t, defaultFS, ta.pick(), "default theme must use default FS")

	withTheme(t, "interapi-unregistered")
	require.Same(t, defaultFS, ta.pick(), "unregistered theme must fall back to default FS")

	// Defensive: nil alternatives map must not panic, default must win.
	nilFS := NewThemeAwareFS(defaultFS, nil)
	nilTA := nilFS.(*themeAwareFileSystem)
	withTheme(t, "classic")
	require.Same(t, defaultFS, nilTA.pick(), "nil alternatives must fall back to default FS")
}

// TestThemeAwareFSDispatchesExistsOpen confirms Exists/Open delegate to
// the picked FS, not just the pick() helper.
func TestThemeAwareFSDispatchesExistsOpen(t *testing.T) {
	called := &stubFS{id: "called", ok: true}
	fs := NewThemeAwareFS(called, nil)
	withTheme(t, "default")
	assert.True(t, fs.Exists("/", "/anything"), "Exists must dispatch to picked FS")
	_, err := fs.Open("/whatever")
	assert.NoError(t, err, "Open must dispatch to picked FS without error")
}
