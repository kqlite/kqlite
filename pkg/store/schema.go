package store

// Kqlite system database for storing replica configuration.
const (
	// Database name for remote DSN.
	ReplicaDB = "kqlite-replicas"

	// Database file for local DSN.
	ReplicaDBFile = "kqlite-replicas.db"

	// Replicas schema/table definitions.
	// addr is in the form host:port or :port
	ReplicasSchema = `
CREATE TABLE IF NOT EXISTS replicas
	(
		id		INTEGER PRIMARY KEY,
		addr	TEXT,
		db		TEXT
	)`
)
