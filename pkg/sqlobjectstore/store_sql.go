package sqlobjectstore

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	// required for sqlite3
	_ "github.com/mattn/go-sqlite3"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/migration"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
)

// nolint: lll
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
	`CREATE INDEX Created_idx ON Objects(Created);`,
	`CREATE INDEX TTL_LastAccessed_idx ON Objects(TTL, LastAccessed);`,
	`CREATE INDEX Type_RootHash_OwnerPublicKey_idx ON Objects(Type, RootHash, OwnerPublicKey);`,
	`CREATE INDEX RootHash_idx ON Objects(RootHash);`,
	`CREATE INDEX RootHash_TTL_idx ON Objects(RootHash, TTL);`,
	`CREATE INDEX Hash_LastAccessed_idx ON Objects(Hash, LastAccessed);`,
	`CREATE TABLE IF NOT EXISTS Relations (Parent TEXT NOT NULL, Child TEXT NOT NULL, PRIMARY KEY (Parent, Child));`,
	`ALTER TABLE Relations ADD RootHash TEXT;`,
	`CREATE INDEX Relations_RootHash_idx ON Relations(RootHash);`,
}

type Store struct {
	db *sql.DB
}

func New(
	db *sql.DB,
) (*Store, error) {
	ndb := &Store{
		db: db,
	}

	// run migrations
	if err := migration.Up(db, migrations...); err != nil {
		return nil, errors.Wrap(err, errors.New("failed to run migrations"))
	}

	// Initialize the garbage collector in the background to run every minute
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
) (*object.Object, error) {
	// get the object
	stmt, err := st.db.Prepare("SELECT Body FROM Objects WHERE Hash=?")
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare query"),
		)
	}
	defer stmt.Close() // nolint: errcheck

	row := stmt.QueryRow(hash.String())

	m := map[string]interface{}{}
	data := []byte{}

	if err := row.Scan(&data); err != nil {
		return nil, errors.Wrap(
			err,
			objectstore.ErrNotFound,
		)
	}

	if err := json.Unmarshal(data, &m); err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not unmarshal data"),
		)
	}

	obj := object.FromMap(m)

	// update the last accessed column
	istmt, err := st.db.Prepare(
		"UPDATE Objects SET LastAccessed=? WHERE Hash=?")
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare query"),
		)
	}
	defer istmt.Close() // nolint: errcheck

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

func (st *Store) GetByStream(
	streamRootHash object.Hash,
) (object.ReadCloser, error) {
	return st.Filter(
		FilterByStreamHash(streamRootHash),
	)
}

func (st *Store) GetByType(
	objectType string,
) (object.ReadCloser, error) {
	return st.Filter(
		FilterByObjectType(objectType),
	)
}

func (st *Store) Put(
	obj *object.Object,
) error {
	return st.PutWithTTL(obj, 0)
}

func (st *Store) PutWithTTL(
	obj *object.Object,
	ttl time.Duration,
) error {
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
	defer stmt.Close() // nolint: errcheck

	body, err := json.Marshal(obj.ToMap())
	if err != nil {
		return errors.Wrap(err, errors.New("could not marshal object"))
	}

	objHash := obj.Hash()
	objectType := obj.Type
	objectHash := objHash.String()
	streamHash := obj.Metadata.Stream.String()
	// TODO support multiple owners
	ownerPublicKey := ""
	if !obj.Metadata.Owner.IsEmpty() {
		ownerPublicKey = obj.Metadata.Owner.String()
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
		ttl,
		time.Now().Unix(),
	)
	if err != nil {
		return errors.Wrap(err, errors.New("could not insert to objects table"))
	}

	for _, p := range obj.Metadata.Parents {
		err := st.putRelation(object.Hash(streamHash), objHash, p)
		if err != nil {
			return errors.Wrap(err, errors.New("could not create relation"))
		}
	}

	return nil
}

func (st *Store) putRelation(
	stream object.Hash,
	parent object.Hash,
	child object.Hash,
) error {
	stmt, err := st.db.Prepare(`
		INSERT OR IGNORE INTO Relations (
			RootHash,
			Parent,
			Child
		) VALUES (
			?, ?, ?
		)
	`)
	if err != nil {
		return errors.Wrap(err,
			errors.New("could not prepare insert to objects table"))
	}
	defer stmt.Close() // nolint: errcheck

	_, err = stmt.Exec(
		stream.String(),
		parent.String(),
		child.String(),
	)
	if err != nil {
		return errors.Wrap(err, errors.New("could not insert to objects table"))
	}

	return nil
}

func (st *Store) GetStreamLeaves(
	streamRootHash object.Hash,
) ([]object.Hash, error) {
	stmt, err := st.db.Prepare(`
		SELECT Parent
		FROM Relations
		WHERE
			RootHash=?
			AND Parent NOT IN (
				SELECT DISTINCT Child
				FROM Relations
				WHERE
					RootHash=?
			)
	`)
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not prepare query"))
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(streamRootHash.String(), streamRootHash.String())
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not query"))
	}
	defer rows.Close() // nolint: errcheck

	hashList := []object.Hash{}

	for rows.Next() {
		data := ""
		if err := rows.Scan(&data); err != nil {
			return nil, errors.Wrap(
				err,
				objectstore.ErrNotFound,
			)
		}
		hashList = append(hashList, object.Hash(data))
	}

	return hashList, nil
}

func (st *Store) GetRelations(
	parent object.Hash,
) ([]object.Hash, error) {
	stmt, err := st.db.Prepare("SELECT Hash FROM Objects WHERE RootHash=?")
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not prepare query"))
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(parent.String())
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not query"))
	}
	defer rows.Close() // nolint: errcheck

	hashList := []object.Hash{}

	for rows.Next() {
		data := ""
		if err := rows.Scan(&data); err != nil {
			return nil, errors.Wrap(
				err,
				objectstore.ErrNotFound,
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
	defer istmt.Close() // nolint: errcheck

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
	defer stmt.Close() // nolint: errcheck

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
	defer stmt.Close() // nolint: errcheck

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

func (st *Store) gc() error {
	stmt, err := st.db.Prepare(`
	DELETE FROM Objects WHERE
	  TTL > 0 AND
	  datetime(LastAccessed + TTL * 60, 'unixepoch') < datetime ('now');
	`)
	if err != nil {
		return errors.Wrap(err, errors.New("could not prepare query"))
	}
	defer stmt.Close() // nolint: errcheck

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
) (object.ReadCloser, error) {
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

	// get the object
	// nolint: gosec
	stmt, err := st.db.Prepare("SELECT Hash FROM Objects " + where)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare statement"),
		)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(whereArgs...)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not query"),
		)
	}
	defer rows.Close() // nolint: errcheck

	hashes := []string{}
	hashesForUpdate := []interface{}{}

	errorChan := make(chan error)
	objectsChan := make(chan *object.Object)
	closeChan := make(chan struct{})

	reader := object.NewReadCloser(
		context.TODO(),
		objectsChan,
		errorChan,
		closeChan,
	)

	for rows.Next() {
		hash := ""
		if err := rows.Scan(&hash); err != nil {
			return nil, errors.Wrap(
				err,
				objectstore.ErrNotFound,
			)
		}
		hashes = append(hashes, hash)
		hashesForUpdate = append(hashesForUpdate, hash)
	}

	if len(hashes) == 0 {
		return nil, objectstore.ErrNotFound
	}

	// update the last accessed column
	updateQs := strings.Repeat(",?", len(hashes))[1:]
	istmt, err := st.db.Prepare(
		"UPDATE Objects SET LastAccessed = ? " +
			"WHERE Hash IN (" + updateQs + ")",
	)
	if err != nil {
		return nil, err
	}
	defer istmt.Close() // nolint: errcheck

	if _, err := istmt.Exec(
		append([]interface{}{time.Now().Unix()}, hashesForUpdate...)...,
	); err != nil {
		return nil, err
	}

	go func() {
		defer close(objectsChan)
		defer close(errorChan)
		for _, hash := range hashes {
			o, err := st.Get(object.Hash(hash))
			if err != nil {
				errorChan <- err
				return
			}
			select {
			case <-closeChan:
				return
			case objectsChan <- o:
				// all good
			}
		}
	}()

	return reader, nil
}

const (
	pinnedQuery = "SELECT Hash FROM Objects WHERE TTL = 0 AND Hash = RootHash"
)

func (st *Store) GetPinned() ([]object.Hash, error) {
	stmt, err := st.db.Prepare(pinnedQuery)
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not prepare statement"),
		)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query()
	if err != nil {
		return nil, errors.Wrap(
			err,
			errors.New("could not query"),
		)
	}
	defer rows.Close() // nolint: errcheck

	hs := []object.Hash{}
	for rows.Next() {
		h := ""
		if err := rows.Scan(&h); err != nil {
			return nil, errors.Wrap(
				err,
				objectstore.ErrNotFound,
			)
		}
		if h != "" {
			hs = append(hs, object.Hash(h))
		}
	}

	return hs, nil
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
