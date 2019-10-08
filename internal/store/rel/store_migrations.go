package rel

var migrations = [...]string{
	`CREATE TABLE IF NOT EXISTS Objects (
	  Hash TEXT NOT NULL PRIMARY KEY,
	  Context TEXT,
	  StreamHash TEXT,
	  Body BLOB
	 )
	`,
}
