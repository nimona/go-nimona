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

const (
	// ErrNotFound is returned when a requested object or hash is not found
	ErrNotFound = errors.Error("not found")
)

const migrationsTable string = `
CREATE TABLE IF NOT EXISTS Migrations (
	ID INTEGER NOT NULL PRIMARY KEY,
	LastIndex INTEGER,
	Datetime INT
)`

type DB struct {
	db *sql.DB
	// the key is the parent object.Hash
	// updates on any put
	subscribers map[string][]chan object.Hash
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
		db:          db,
		subscribers: map[string][]chan object.Hash{},
	}

	if err := ndb.createMigrationTable(); err != nil {
		return nil, err
	}
	if err := ndb.runMigrations(); err != nil {
		return nil, err
	}

	return ndb, nil
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

		rows, err := tx.Query(
			"select ID, LastIndex, Datetime from Migrations order by id desc limit 1")
		if err != nil {
			tx.Rollback() //nolint
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

		_, err = tx.Exec(mig)
		if err != nil {
			tx.Rollback() //nolint
			return errors.Wrap(
				err,
				errors.New("could not run migration"),
			)
		}

		stmt, err := tx.Prepare(
			"INSERT INTO Migrations(LastIndex, Datetime) VALUES(?, ?)")
		if err != nil {
			tx.Rollback() //nolint
			return errors.Wrap(
				err,
				errors.New("could not insert to migrations table"),
			)
		}

		_, err = stmt.Exec(index, time.Now().Unix())
		if err != nil {
			tx.Rollback() //nolint
			return errors.Wrap(
				err,
				errors.New("could not insert to migrations table"),
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(
			err,
			errors.New("could not insert to migrations table"),
		)
	}

	return nil
}

func (d *DB) Get(
	hash object.Hash,
) (object.Object, error) {

	stmt, err := d.db.Prepare("SELECT Body FROM Objects WHERE Hash=?")
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare query"),
		)
	}

	row := stmt.QueryRow(hash.String())

	obj := object.New()
	data := []byte{}

	if err := row.Scan(&data); err != nil {
		return nil, errors.Wrap(
			err,
			ErrNotFound,
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

	if _, err := istmt.Exec(
		time.Now().Unix(),
		hash.String(),
	); err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not update last access"),
		)
	}

	return obj, nil
}

func (d *DB) Put(
	obj object.Object,
	opts ...Option,
) error {
	options := &Options{
		TTL: 0,
	}
	for _, opt := range opts {
		opt(options)
	}

	stmt, err := d.db.Prepare(`
	REPLACE INTO Objects(
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

	stHash := stream.Stream(obj)
	if stHash != nil {
		stHash.String()
	}

	streamHashStr := stHash.String()
	objectHash := hash.New(obj)

	_, err = stmt.Exec(
		objectHash.String(),
		streamHashStr,
		body,
		time.Now().Unix(),
		time.Now().Unix(),
		options.TTL,
	)
	if err != nil {
		return errors.Wrap(err, errors.New("could not insert to objects table"))
	}

	if subs, ok := d.subscribers[streamHashStr]; ok {
		for _, subCh := range subs {
			subCh <- *objectHash
		}
	}
	return nil
}

func (d *DB) GetRelations(
	parent object.Hash,
) ([]*object.Hash, error) {

	stmt, err := d.db.Prepare("SELECT Hash FROM Objects WHERE StreamHash=?")
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not prepare query"))
	}

	rows, err := stmt.Query(parent.String())
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not query"))
	}

	hashList := []*object.Hash{}

	for rows.Next() {

		data := []byte{}
		if err := rows.Scan(&data); err != nil {
			return nil, errors.Wrap(
				err,
				ErrNotFound,
			)
		}

		h, err := object.HashFromCompact(string(data))
		if err != nil {
			continue
		}

		hashList = append(hashList, h)
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

	if _, err := istmt.Exec(
		time.Now().Unix(),
		parent.String(),
	); err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not update last access"),
		)
	}

	return hashList, nil
}

func (d *DB) UpdateTTL(
	hash object.Hash,
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
		hash.String(),
	); err != nil {
		return errors.Wrap(
			err,
			errors.New("could not update last access and ttl"),
		)
	}

	return nil
}

func (d *DB) Delete(
	hash object.Hash,
) error {

	stmt, err := d.db.Prepare(`
	DELETE FROM Objects
	WHERE Hash=?`)
	if err != nil {
		return errors.Wrap(err, errors.New("could not prepare query"))
	}

	if _, err := stmt.Exec(
		hash.String(),
	); err != nil {
		return errors.Wrap(
			err,
			errors.New("could not delete object"),
		)
	}

	return nil
}

func (d *DB) Subscribe(parent object.Hash) (chan object.Hash, error) {
	ch := make(chan object.Hash)
	if _, ok := d.subscribers[parent.String()]; !ok {
		d.subscribers[parent.String()] = []chan object.Hash{}
	}

	d.subscribers[parent.String()] = append(d.subscribers[parent.String()], ch)
	return ch, nil
}
