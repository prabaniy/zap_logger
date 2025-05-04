package main

import (
	"sync"

	"go.uber.org/zap/zapcore"
)

// multiCoreSyncWrapper wraps multiple zapcore.Core implementations
// and provides thread-safe access to the collection
type multiCoreSyncWrapper struct {
	cores []zapcore.Core
	mu    sync.RWMutex
}

// Enabled implements zapcore.Core
func (m *multiCoreSyncWrapper) Enabled(lvl zapcore.Level) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, core := range m.cores {
		if core.Enabled(lvl) {
			return true
		}
	}
	return false
}

// With implements zapcore.Core
func (m *multiCoreSyncWrapper) With(fields []zapcore.Field) zapcore.Core {
	m.mu.Lock()
	defer m.mu.Unlock()

	cores := make([]zapcore.Core, 0, len(m.cores))
	for _, core := range m.cores {
		cores = append(cores, core.With(fields))
	}

	return &multiCoreSyncWrapper{cores: cores}
}

// Check implements zapcore.Core
func (m *multiCoreSyncWrapper) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, core := range m.cores {
		ce = core.Check(ent, ce)
	}
	return ce
}

// Write implements zapcore.Core
func (m *multiCoreSyncWrapper) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, core := range m.cores {
		if err := core.Write(ent, fields); err != nil {
			return err
		}
	}
	return nil
}

// Sync implements zapcore.Core
func (m *multiCoreSyncWrapper) Sync() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, core := range m.cores {
		if err := core.Sync(); err != nil {
			return err
		}
	}
	return nil
}

// AddCore adds a new zapcore.Core to the wrapper
func (m *multiCoreSyncWrapper) AddCore(core zapcore.Core) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cores = append(m.cores, core)
}
