package sql

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/object"
	"nimona.io/pkg/stream"
)

//go:generate $GOBIN/genny -in=$GENERATORS/pubsub/pubsub.go -out=pubsub.go -pkg sql gen "ObjectType=object.Object PubSubName=sqlStore"

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

type Store struct {
	db     *sql.DB
	pubsub SqlStorePubSub
}

type migrationRow struct {
	id        int
	LastIndex int
	Datetime  string
}

func New(
	db *sql.DB,
) (*Store, error) {
	ndb := &Store{
		db:     db,
		pubsub: NewSqlStorePubSub(),
	}

	// initialise the tables required for the migration
	if err := ndb.createMigrationTable(); err != nil {
		return nil, err
	}

	// Execute the migrations
	if err := ndb.runMigrations(); err != nil {
		return nil, err
	}

	// Initialise the garbage collector in the background to run every minute
	go func() {
		for {
			ndb.gc() // nolint
			time.Sleep(1 * time.Minute)
		}
	}()

	return ndb, nil
}

func (st *Store) Close() error {
	return st.db.Close()
}

// createMigrationTable creates the tables required to keep the state
// of the migrations
func (st *Store) createMigrationTable() error {
	_, err := st.db.Exec(migrationsTable)
	if err != nil {
		return errors.Wrap(err, errors.New("could not create migrations table"))
	}

	return nil
}

// runMigrations executes the migrations in the array and stores the state
// in the migration tables
func (st *Store) runMigrations() error {
	tx, err := st.db.Begin()
	if err != nil {
		return errors.Wrap(err, errors.New("could not start transaction"))
	}

	// iterate over the migrations array
	for index, mig := range migrations {

		// get the last migration index
		rows, err := tx.Query(
			"select ID, LastIndex, Datetime from Migrations order by id desc limit 1")
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

func (st *Store) Get(
	hash object.Hash,
) (object.Object, error) {
	// get the object
	stmt, err := st.db.Prepare("SELECT Body FROM Objects WHERE Hash=?")
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

	// update the last accessed column
	istmt, err := st.db.Prepare(
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

func (st *Store) Put(
	obj object.Object,
	opts ...Option,
) error {
	options := &Options{
		TTL: 0,
	}
	for _, opt := range opts {
		opt(options)
	}

	stmt, err := st.db.Prepare(`
	REPLACE INTO Objects (
		Hash,
		Type,
		RootHash,
		Body,
		Created,
		LastAccessed,
		TTl
	) VALUES (
		?, ?, ?, ?, ?, ?, ?
	) ON CONFLICT (Hash) DO UPDATE SET
		LastAccessed=?
	`)
	if err != nil {
		return errors.Wrap(err,
			errors.New("could not prepare insert to objects table"))
	}

	body, err := json.Marshal(obj.ToMap())
	if err != nil {
		return errors.Wrap(err, errors.New("could not marshal object"))
	}

	objectType := obj.GetType()
	streamHash := stream.Stream(obj).String()
	objectHash := hash.New(obj).String()

	// if the object doesn't belong to a stream, we need to set the stream
	// to the object's hash.
	// This should allow queries to consider the root object part of the stream.
	if streamHash == "" {
		streamHash = objectHash
	}

	_, err = stmt.Exec(
		objectHash,
		objectType,
		streamHash,
		body,
		time.Now().Unix(),
		time.Now().Unix(),
		options.TTL,
		time.Now().Unix(),
	)
	if err != nil {
		return errors.Wrap(err, errors.New("could not insert to objects table"))
	}

	st.pubsub.Publish(obj)

	return nil
}

func (st *Store) GetRelations(
	parent object.Hash,
) ([]object.Hash, error) {
	stmt, err := st.db.Prepare("SELECT Hash FROM Objects WHERE RootHash=?")
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not prepare query"))
	}

	rows, err := stmt.Query(parent.String())
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not query"))
	}

	hashList := []object.Hash{}

	for rows.Next() {
		data := ""
		if err := rows.Scan(&data); err != nil {
			return nil, errors.Wrap(
				err,
				ErrNotFound,
			)
		}
		hashList = append(hashList, object.Hash(data))
	}

	istmt, err := st.db.Prepare(
		"UPDATE Objects SET LastAccessed=? WHERE RootHash=?",
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

func (st *Store) UpdateTTL(
	hash object.Hash,
	minutes int,
) error {
	stmt, err := st.db.Prepare(`UPDATE Objects SET TTL=? WHERE RootHash=?`)
	if err != nil {
		return errors.Wrap(err, errors.New("could not prepare query"))
	}

	if _, err := stmt.Exec(minutes, hash.String()); err != nil {
		return errors.Wrap(
			err,
			errors.New("could not update last access and ttl"),
		)
	}

	return nil
}

func (st *Store) Remove(
	hash object.Hash,
) error {
	stmt, err := st.db.Prepare(`
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

func (st *Store) Subscribe(
	lookupOptions ...LookupOption,
) SqlStoreSubscription {
	options := &LookupOptions{}
	for _, lookupOption := range lookupOptions {
		lookupOption(options)
	}
	return st.pubsub.Subscribe(options.Filters...)
}

func (st *Store) gc() error {
	stmt, err := st.db.Prepare(`
	DELETE
	FROM
		Objects
	WHERE
		datetime (LastAccessed + TTL * 60, 'unixepoch') < datetime ('now');
	`)
	if err != nil {
		return errors.Wrap(err, errors.New("could not prepare query"))
	}

	if _, err := stmt.Exec(); err != nil {
		return errors.Wrap(
			err,
			errors.New("could not gc delete objects"),
		)
	}

	return nil
}

func (st *Store) Filter(
	lookupOptions ...LookupOption,
) ([]object.Object, error) {
	options := &LookupOptions{}
	for _, lookupOption := range lookupOptions {
		lookupOption(options)
	}

	where := "WHERE 1 "
	whereArgs := []interface{}{}

	if len(options.Lookups.ObjectHashes) > 0 {
		qs := strings.Repeat(",?", len(options.Lookups.ObjectHashes))[1:]
		where += "AND Hash IN (" + qs + ") "
		whereArgs = append(whereArgs, ahtoai(options.Lookups.ObjectHashes)...)
	}

	if len(options.Lookups.ContentTypes) > 0 {
		qs := strings.Repeat(",?", len(options.Lookups.ContentTypes))[1:]
		where += "AND Type IN (" + qs + ") "
		whereArgs = append(whereArgs, astoai(options.Lookups.ContentTypes)...)
	}

	if len(options.Lookups.StreamHashes) > 0 {
		qs := strings.Repeat(",?", len(options.Lookups.StreamHashes))[1:]
		where += "AND RootHash IN (" + qs + ") "
		whereArgs = append(whereArgs, ahtoai(options.Lookups.StreamHashes)...)
	}

	objects := []object.Object{}

	// get the object
	stmt, err := st.db.Prepare("SELECT Body FROM Objects " + where)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare where"),
		)
	}

	rows, err := stmt.Query(whereArgs...)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not query"),
		)
	}

	hashes := []interface{}{}

	for rows.Next() {
		obj := object.New()
		data := []byte{}

		if err := rows.Scan(&data); err != nil {
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

		objects = append(objects, obj)
		hashes = append(hashes, hash.New(obj))
	}

	// update the last accessed column
	updateQs := strings.Repeat(",?", len(hashes))[1:]
	istmt, err := st.db.Prepare(
		"UPDATE Objects SET LastAccessed = ? " +
			"WHERE Hash IN (" + updateQs + ")",
	)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare query"),
		)
	}

	if _, err := istmt.Exec(
		append([]interface{}{time.Now().Unix()}, hashes...)...,
	); err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not update last access"),
		)
	}

	return objects, nil
}

func astoai(ah []string) []interface{} {
	as := make([]interface{}, len(ah))
	for i, h := range ah {
		as[i] = h
	}
	return as
}

func ahtoai(ah []object.Hash) []interface{} {
	as := make([]interface{}, len(ah))
	for i, h := range ah {
		as[i] = h.String()
	}
	return as
}
