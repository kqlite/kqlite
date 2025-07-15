package sysdb

const SystemSchema = `
CREATE TABLE IF NOT EXISTS replica
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		host TEXT,
		port INTEGER,
		dbname TEXT,
		synced INT8
	)`
