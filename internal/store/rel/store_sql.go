package rel

import (
	"database/sql"
	"encoding/json"
	"time"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/object"
	"nimona.io/pkg/stream"
)

const migrationsTable string = `
CREATE TABLE IF NOT EXISTS Migrations (
	ID INTEGER NOT NULL PRIMARY KEY,
	LastIndex INTEGER,
	Datetime INT
)`

type DB struct {
	db *sql.DB
}

type migrationRow struct {
	id        int
	LastIndex int
	Datetime  string
}

func New(
	db *sql.DB,
) (*DB, error) {
	ndb := &DB{
		db: db,
	}

	err := ndb.createMigrationTable()
	err = ndb.runMigrations()

	return ndb, err
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) createMigrationTable() error {
	_, err := d.db.Exec(migrationsTable)
	if err != nil {
		return errors.Wrap(err, errors.New("could not create migrations table"))
	}

	return nil
}

func (d *DB) runMigrations() error {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, errors.New("could not start transaction"))
	}

	for index, mig := range migrations {

		rows, err := tx.Query("select ID, LastIndex, Datetime from Migrations order by id desc limit 1")
		if err != nil {
			tx.Rollback()
			return errors.Wrap(
				err,
				errors.New("could not run migration"),
			)
		}

		mgr := migrationRow{}

		for rows.Next() {
			rows.Scan(&mgr.id, &mgr.LastIndex, &mgr.Datetime)
		}

		if mgr.id > 0 && mgr.LastIndex >= index {
			continue
		}

		_, err = tx.Exec(mig)
		if err != nil {
			tx.Rollback()
			return errors.Wrap(
				err,
				errors.New("could not run migration"),
			)
		}

		stmt, err := tx.Prepare(
			"INSERT INTO Migrations(LastIndex, Datetime) VALUES(?, ?)")
		if err != nil {
			tx.Rollback()
			return errors.Wrap(
				err,
				errors.New("could not insert to migrations table"),
			)
		}

		_, err = stmt.Exec(index, time.Now().Unix())
		if err != nil {
			tx.Rollback()
			return errors.Wrap(
				err,
				errors.New("could not insert to migrations table"),
			)
		}
	}

	tx.Commit()

	return nil
}

func (d *DB) GetByHash(
	hash string,
) (object.Object, error) {

	stmt, err := d.db.Prepare("SELECT Body FROM Objects WHERE Hash=?")
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare query"),
		)
	}

	row := stmt.QueryRow(hash)

	obj := object.New()
	data := []byte{}

	if err := row.Scan(&data); err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not query objects"),
		)
	}

	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not unmarshal data"),
		)
	}

	istmt, err := d.db.Prepare(
		"UPDATE Objects SET LastAccessed=? WHERE Hash=?")
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare query"),
		)
	}

	if _, err := istmt.Exec(time.Now().Unix(), hash); err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not update last access"),
		)
	}

	return obj, nil
}

func (d *DB) Store(
	obj object.Object,
	ttl int, // minutes
) error {
	stmt, err := d.db.Prepare(`
	INSERT INTO Objects(
		Hash,
		StreamHash,
		Body,
		Created,
		LastAccessed,
		TTl
	)
	VALUES(?, ?, ?, ?, ? ,?)
	`)
	if err != nil {
		return errors.Wrap(err,
			errors.New("could not prepare insert to objects table"))
	}

	body, err := json.Marshal(obj.ToMap())
	if err != nil {
		return errors.Wrap(err, errors.New("could not marshal object"))
	}

	stHashStr := ""
	stHash := stream.Stream(obj)
	if stHash != nil {
		stHash.String()
	}

	_, err = stmt.Exec(
		hash.New(obj).String(),
		stHashStr,
		body,
		time.Now().Unix(),
		time.Now().Unix(),
		ttl,
	)
	if err != nil {
		return errors.Wrap(err, errors.New("could not insert to objects table"))
	}

	return nil
}

func (d *DB) GetByStreamHash(
	streamHash string,
) (object.Object, error) {

	stmt, err := d.db.Prepare("SELECT Body FROM Objects WHERE StreamHash=?")
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not prepare query"))
	}

	row := stmt.QueryRow(streamHash)

	obj := object.New()
	data := []byte{}

	if err := row.Scan(&data); err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not query objects"),
		)
	}

	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not unmarshal data"),
		)
	}

	istmt, err := d.db.Prepare(
		"UPDATE Objects SET LastAccessed=? WHERE StreamHash=?",
	)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare query"),
		)
	}

	if _, err := istmt.Exec(time.Now().Unix(), streamHash); err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not update last access"),
		)
	}

	return obj, nil
}

func (d *DB) UpdateTTL(
	hash string,
	minutes int,
) error {

	stmt, err := d.db.Prepare(`
	UPDATE Objects
	SET TTL=?, LastAccessed=?
	WHERE Hash=?`)
	if err != nil {
		return errors.Wrap(err, errors.New("could not prepare query"))
	}

	if _, err := stmt.Exec(
		minutes,
		time.Now().Unix(),
		hash,
	); err != nil {
		return errors.Wrap(
			err,
			errors.New("could not update last access and ttl"),
		)
	}

	return nil
}
