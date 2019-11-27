package sql

var migrations = [...]string{
	`CREATE TABLE IF NOT EXISTS Objects (
	  Hash TEXT NOT NULL PRIMARY KEY,
	  Type TEXT,
	  StreamHash TEXT,
	  Body BLOB,
	  Created INT,
	  LastAccessed INT,
	  TTL INT
	 )
	`,
}
