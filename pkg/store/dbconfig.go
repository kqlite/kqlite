package store

import "path/filepath"

// DBConfig represents the configuration of the underlying SQLite database.
type DBConfig struct {
	// SQLite on-disk path
	OnDiskPath string

	// Enforce Foreign Key constraints
	FKConstraints bool

	// Paths of SQLite Extensions to be loaded
	Extensions []string

	// Open the SQLite datbase in read-only mode.
	ReadOnly bool

	// Open the SQLite datbase in WAL mode.
	WalEnabled bool
}

// NewDBConfig returns a new DB config instance.
func NewDBConfig() *DBConfig {
	return &DBConfig{}
}

// ExtensionNames returns the names of the SQLite extensions.
func (c *DBConfig) ExtensionNames() []string {
	names := make([]string, 0, len(c.Extensions))
	for _, ext := range c.Extensions {
		names = append(names, filepath.Base(ext))
	}
	return names
}
