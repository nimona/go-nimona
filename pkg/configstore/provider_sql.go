package configstore

import (
	"database/sql"
	"fmt"

	// required for sqlite3
	_ "modernc.org/sqlite"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/migration"
	"nimona.io/pkg/objectstore"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS Configs (Key TEXT NOT NULL PRIMARY KEY);`,
	`ALTER TABLE Configs ADD Value TEXT;`,
}

type SQLProvider struct {
	db *sql.DB
}

func NewSQLProvider(
	db *sql.DB,
) (*SQLProvider, error) {
	p := &SQLProvider{
		db: db,
	}

	// run migrations
	if err := migration.Up(db, migrations...); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return p, nil
}

func (p *SQLProvider) Close() error {
	return p.db.Close()
}

func (p *SQLProvider) Put(
	key string,
	value string,
) error {
	stmt, err := p.db.Prepare(`
		INSERT OR REPLACE INTO Configs (Key, Value) VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("could not prepare insert to configs, %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	_, err = stmt.Exec(
		key,
		value,
	)
	if err != nil {
		return fmt.Errorf("could not insert to configs table, %w", err)
	}

	return nil
}

func (p *SQLProvider) Get(key string) (string, error) {
	stmt, err := p.db.Prepare(`
		SELECT Value FROM Configs WHERE Key = ?
	`)
	if err != nil {
		return "", fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(key)
	if err != nil {
		return "", fmt.Errorf("could not query: %w", err)
	}
	defer rows.Close() // nolint: errcheck

	v := ""
	rows.Next()
	if err := rows.Scan(&v); err != nil {
		return "", errors.Merge(objectstore.ErrNotFound, err)
	}

	return v, nil
}

func (p *SQLProvider) List() (map[string]string, error) {
	stmt, err := p.db.Prepare(`
		SELECT Key, Value FROM Configs
	`)
	if err != nil {
		return nil, fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("could not query: %w", err)
	}
	defer rows.Close() // nolint: errcheck

	cfg := map[string]string{}
	for rows.Next() {
		k := ""
		v := ""
		if err := rows.Scan(&k, &v); err != nil {
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		if k != "" {
			cfg[k] = v
		}
	}

	return cfg, nil
}

func (p *SQLProvider) Remove(key string) error {
	stmt, err := p.db.Prepare(`
		DELETE FROM Configs WHERE Key = ?
	`)
	if err != nil {
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	if _, err := stmt.Exec(key); err != nil {
		return fmt.Errorf("could not query: %w", err)
	}

	return nil
}
