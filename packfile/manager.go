package packfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

/*

	packfile manager keeps track of open file descriptors of packfiles
	and caches the fd to avoid opening and closing files for every request.
*/

type Manager struct {
	dir string

	mu    sync.RWMutex
	files map[uint64]*os.File
}

func NewManager(dir string) *Manager {
	return &Manager{
		dir:   dir,
		files: make(map[uint64]*os.File),
	}
}

func (m *Manager) path(packID uint64) string {
	name := fmt.Sprintf("%016x.pack", packID)
	return filepath.Join(m.dir, name)
}

func (m *Manager) Open(packID uint64) (*os.File, error) {
	m.mu.RLock()
	f, ok := m.files[packID]
	m.mu.RUnlock()

	if ok {
		return f, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Another goroutine may have opened it while we waited.
	if f, ok := m.files[packID]; ok {
		return f, nil
	}

	path := m.path(packID)

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("packfile: open pack %d: %w", packID, err)
	}

	m.files[packID] = f

	return f, nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error

	for id, f := range m.files {
		if err := f.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("packfile: close pack %d: %w", id, err)
		}
	}

	m.files = make(map[uint64]*os.File)

	return firstErr
}
