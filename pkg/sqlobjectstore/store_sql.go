package sqlobjectstore

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/migration"
	"nimona.io/pkg/object"
)

//go:generate $GOBIN/genny -in=$GENERATORS/pubsub/pubsub.go -out=pubsub_generated.go -pkg sqlobjectstore gen "ObjectType=object.Object PubSubName=sqlStore"

const (
	// ErrNotFound is returned when a requested object or hash is not found
	ErrNotFound = errors.Error("not found")
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS Objects (Hash TEXT NOT NULL PRIMARY KEY);`,
	`ALTER TABLE Objects ADD Type TEXT;`,
	`ALTER TABLE Objects ADD Body TEXT;`,
	`ALTER TABLE Objects ADD RootHash TEXT;`,
	`ALTER TABLE Objects ADD TTL INT;`,
	`ALTER TABLE Objects ADD Created INT;`,
	`ALTER TABLE Objects ADD LastAccessed INT;`,
	`ALTER TABLE Objects ADD SignerPublicKey TEXT;`,
	`ALTER TABLE Objects ADD AuthorPublicKey TEXT;`,
	`ALTER TABLE Objects RENAME AuthorPublicKey TO OwnerPublicKey;`,
	`ALTER TABLE Objects RENAME SignerPublicKey TO _DeprecatedSignerPublicKey;`,
}

type Store struct {
	db     *sql.DB
	pubsub SqlStorePubSub
}

func New(
	db *sql.DB,
) (*Store, error) {
	ndb := &Store{
		db:     db,
		pubsub: NewSqlStorePubSub(),
	}

	// run migrations
	if err := migration.Up(db, migrations...); err != nil {
		return nil, errors.Wrap(err, errors.New("failed to run migrations"))
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

func (st *Store) Get(
	hash object.Hash,
) (object.Object, error) {
	obj := object.Object{}

	// get the object
	stmt, err := st.db.Prepare("SELECT Body FROM Objects WHERE Hash=?")
	if err != nil {
		return obj, errors.Wrap(
			err,
			errors.New("could not prepare query"),
		)
	}

	row := stmt.QueryRow(hash.String())

	m := map[string]interface{}{}
	data := []byte{}

	if err := row.Scan(&data); err != nil {
		return obj, errors.Wrap(
			err,
			ErrNotFound,
		)
	}

	if err := json.Unmarshal(data, &m); err != nil {
		return obj, errors.Wrap(
			err,
			errors.New("could not unmarshal data"),
		)
	}

	obj = object.FromMap(m)

	// update the last accessed column
	istmt, err := st.db.Prepare(
		"UPDATE Objects SET LastAccessed=? WHERE Hash=?")
	if err != nil {
		return obj, errors.Wrap(
			err,
			errors.New("could not prepare query"),
		)
	}

	if _, err := istmt.Exec(
		time.Now().Unix(),
		hash.String(),
	); err != nil {
		return obj, errors.Wrap(
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

	// TODO(geoah) why replace?
	stmt, err := st.db.Prepare(`
	REPLACE INTO Objects (
		Hash,
		Type,
		RootHash,
		OwnerPublicKey,
		Body,
		Created,
		LastAccessed,
		TTl
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?
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
	objectHash := object.NewHash(obj).String()
	streamHash := obj.GetStream().String()
	// TODO support multiple owners
	ownerPublicKey := ""
	if len(obj.GetOwners()) > 0 {
		ownerPublicKey = obj.GetOwners()[0].String()
	}

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
		ownerPublicKey,
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
	options := newLookupOptions(lookupOptions...)
	ps := st.pubsub
	return ps.Subscribe(options.Filters...)
}

func (st *Store) gc() error {
	stmt, err := st.db.Prepare(`
	DELETE FROM Objects WHERE
	  TTL > 0 AND
	  datetime(LastAccessed + TTL * 60, 'unixepoch') < datetime ('now');
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
	options := newLookupOptions(lookupOptions...)

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

	if len(options.Lookups.Owners) > 0 {
		qs := strings.Repeat(",?", len(options.Lookups.Owners))[1:]
		where += "AND OwnerPublicKey IN (" + qs + ") "
		whereArgs = append(whereArgs, aktoai(options.Lookups.Owners)...)
	}

	where += "ORDER BY Created ASC"

	objects := []object.Object{}

	// get the object
	stmt, err := st.db.Prepare("SELECT Body FROM Objects " + where)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare statement"),
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
		data := []byte{}

		if err := rows.Scan(&data); err != nil {
			return nil, errors.Wrap(
				err,
				ErrNotFound,
			)
		}

		m := map[string]interface{}{}
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, errors.Wrap(
				err,
				errors.New("could not unmarshal data"),
			)
		}

		obj := object.FromMap(m)
		objects = append(objects, obj)
		hashes = append(hashes, object.NewHash(obj))
	}

	if len(hashes) == 0 {
		return objects, nil
	}

	// update the last accessed column
	updateQs := strings.Repeat(",?", len(hashes))[1:]
	istmt, err := st.db.Prepare(
		"UPDATE Objects SET LastAccessed = ? " +
			"WHERE Hash IN (" + updateQs + ")",
	)
	if err != nil {
		return objects, nil
		// return nil, errors.Wrap(
		// 	err,
		// 	errors.New("could not prepare query"),
		// )
	}

	if _, err := istmt.Exec(
		append([]interface{}{time.Now().Unix()}, hashes...)...,
	); err != nil {
		return objects, nil
		// return nil, errors.Wrap(
		// 	err,
		// 	errors.New("could not update last access"),
		// )
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

func aktoai(ah []crypto.PublicKey) []interface{} {
	as := make([]interface{}, len(ah))
	for i, h := range ah {
		as[i] = h.String()
	}
	return as
}
