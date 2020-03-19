package migration

import (
	"database/sql"
	"time"

	"nimona.io/pkg/errors"
)

const migrationsTable string = `
	CREATE TABLE IF NOT EXISTS Migrations (
		ID INTEGER NOT NULL PRIMARY KEY,
		LastIndex INTEGER,
		Datetime INT
	);`

type migrationRow struct {
	id        int
	LastIndex int
	Datetime  string
}

func Up(db *sql.DB, migrations ...string) error {
	// initialize the tables required for the migration
	if err := createMigrationTable(db); err != nil {
		return err
	}

	// Execute the migrations
	if err := migrateUp(db, migrations...); err != nil {
		return err
	}

	return nil
}

// createMigrationTable creates the tables required to keep the state
// of the migrations
func createMigrationTable(db *sql.DB) error {
	_, err := db.Exec(migrationsTable)
	if err != nil {
		return errors.Wrap(err, errors.New("could not create migrations table"))
	}

	return nil
}

// migrateUp etxecutes the migrations in the array and stores the state
// in the migration tables
func migrateUp(db *sql.DB, migrations ...string) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, errors.New("could not start transaction"))
	}

	// iterate over the migrations array
	for index, mig := range migrations {
		// get the last migration index
		rows, err := tx.Query(`
			SELECT 
				ID,
				LastIndex,
				Datetime
			FROM Migrations
			ORDER BY ID DESC
			LIMIT 1;
		`)
		if err != nil {
			tx.Rollback() // nolint
			return errors.Wrap(
				err,
				errors.New("could not run migration"),
			)
		}

		mgr := migrationRow{}

		for rows.Next() {
			if err := rows.Scan(&mgr.id, &mgr.LastIndex, &mgr.Datetime); err != nil {
				continue
			}
		}

		if mgr.id > 0 && mgr.LastIndex >= index {
			continue
		}

		// execute the current migration
		_, err = tx.Exec(mig)
		if err != nil {
			tx.Rollback() // nolint
			return errors.Wrap(
				err,
				errors.New("could not run migration"),
			)
		}

		// store the migration status state in the table
		stmt, err := tx.Prepare(
			"INSERT INTO Migrations(LastIndex, Datetime) VALUES(?, ?)")
		if err != nil {
			tx.Rollback() // nolint
			return errors.Wrap(
				err,
				errors.New("could not insert to migrations table"),
			)
		}

		_, err = stmt.Exec(index, time.Now().Unix())
		if err != nil {
			tx.Rollback() // nolint
			return errors.Wrap(
				err,
				errors.New("could not insert to migrations table"),
			)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback() // nolint
		return errors.Wrap(
			err,
			errors.New("could not insert to migrations table"),
		)
	}

	return nil
}
