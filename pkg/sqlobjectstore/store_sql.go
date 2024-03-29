package sqlobjectstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	// required for sqlite3
	_ "modernc.org/sqlite"

	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/migration"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/tilde"
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
	`ALTER TABLE Objects ADD MetadataDatetime INT DEFAULT 0;`,
	`CREATE TABLE IF NOT EXISTS Pins (Hash TEXT NOT NULL PRIMARY KEY);`,
	`CREATE TABLE IF NOT EXISTS Keys (PublicKeyDigest TEXT NOT NULL PRIMARY KEY);`,
	`ALTER TABLE Keys ADD PrivateKey TEXT;`,
	`ALTER TABLE Objects ADD Sequence INT;`,
}

var defaultTTL = time.Hour * 24 * 7

// Note(geoah): Sqlite is a bit iffy when trying to write while something is
// reading, ie using rows.Next and results in db lock errors. For this reason
// a mutex for each table has been added
// https://github.com/mattn/go-sqlite3/issues/607#issuecomment-808739698

// TODO: Remove completely LastAccessed or move them to a different table.

type (
	Store struct {
		db               *sql.DB
		listeners        map[string]chan Event
		listenersLock    sync.RWMutex
		tableLockObjects sync.Mutex
		tableLockPins    sync.Mutex
		tableLockKeys    sync.Mutex
	}
	EventAction string
	Event       struct {
		Action     EventAction
		ObjectHash tilde.Digest
	}
)

const (
	ObjectInserted EventAction = "objectInserted"
	ObjectRemoved  EventAction = "objectRemoved"
	ObjectPinned   EventAction = "objectPinned"
	ObjectUnpinned EventAction = "objectUnpinned"
)

func New(
	db *sql.DB,
) (*Store, error) {
	ndb := &Store{
		db:               db,
		listeners:        map[string]chan Event{},
		listenersLock:    sync.RWMutex{},
		tableLockObjects: sync.Mutex{},
		tableLockPins:    sync.Mutex{},
		tableLockKeys:    sync.Mutex{},
	}

	// run migrations
	if err := migration.Up(db, migrations...); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// set pragmas

	_, err := db.Exec("PRAGMA busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("error setting pragmas, %w", err)
	}

	// and verify they were set
	actualPragmaBusyTimeout := 0
	row := db.QueryRow("PRAGMA busy_timeout")
	row.Scan(&actualPragmaBusyTimeout) // nolint: errcheck
	if actualPragmaBusyTimeout != 5000 {
		return nil, fmt.Errorf("unable to set busy_timeout pragma")
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
	hash tilde.Digest,
) (*object.Object, error) {
	st.tableLockObjects.Lock()
	defer st.tableLockObjects.Unlock()

	// get the object
	stmt, err := st.db.Prepare("SELECT Body FROM Objects WHERE Hash=?")
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	row := stmt.QueryRow(hash.String())

	obj := &object.Object{}
	data := []byte{}

	if err := row.Scan(&data); err != nil {
		return nil, errors.Merge(objectstore.ErrNotFound, err)
	}

	if err := json.Unmarshal(data, obj); err != nil {
		return nil, fmt.Errorf("could not unmarshal data: %w", err)
	}

	return obj, nil
}

func (st *Store) GetByStream(
	streamRootHash tilde.Digest,
) (object.ReadCloser, error) {
	return st.Filter(
		FilterByStreamHash(streamRootHash),
		FilterOrderBy("sequence"),
		FilterOrderDir("ASC"),
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
	return st.PutWithTTL(obj, defaultTTL)
}

func (st *Store) PutWithTTL(
	obj *object.Object,
	ttl time.Duration,
) error {
	st.tableLockObjects.Lock()
	defer st.tableLockObjects.Unlock()

	// TODO(geoah) why replace?
	stmt, err := st.db.Prepare(`
	REPLACE INTO Objects (
		Hash,
		Type,
		RootHash,
		Sequence,
		OwnerPublicKey,
		Body,
		Created,
		LastAccessed,
		TTL,
		MetadataDatetime
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	) ON CONFLICT (Hash) DO UPDATE SET
		LastAccessed=?
	`)
	if err != nil {
		return fmt.Errorf("could not prepare insert to objects table: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	body, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("could not marshal object: %w", err)
	}

	objHash := obj.Hash()
	objectType := obj.Type
	objectHash := objHash.String()
	streamHash := obj.Metadata.Root.String()
	// TODO support multiple owners
	ownerPublicKey := ""
	if !obj.Metadata.Owner.IsEmpty() {
		// nolint: errcheck
		ownerPublicKey, _ = obj.Metadata.Owner.MarshalString()
	}

	// if the object doesn't belong to a stream, we need to set the stream
	// to the object's hash.
	// This should allow queries to consider the root object part of the stream.
	if streamHash == "" {
		streamHash = objectHash
	}

	un := 0
	dt, err := time.Parse(
		time.RFC3339,
		obj.Metadata.Timestamp,
	)
	if err == nil {
		un = int(dt.Unix())
	}

	_, err = stmt.Exec(
		// VALUES
		objectHash,
		objectType,
		streamHash,
		obj.Metadata.Sequence,
		ownerPublicKey,
		body,
		time.Now().Unix(),
		time.Now().Unix(),
		int64(ttl.Seconds()),
		un,
		// WHERE
		time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("could not insert to objects table: %w", err)
	}

	if len(obj.Metadata.Parents) > 0 {
		for _, group := range obj.Metadata.Parents {
			for _, p := range group {
				err := st.putRelation(tilde.Digest(streamHash), objHash, p)
				if err != nil {
					return fmt.Errorf("could not create relation: %w", err)
				}
			}
		}
	}

	if streamHash == objectHash {
		err := st.putRelation(tilde.Digest(streamHash), objHash, tilde.EmptyDigest)
		if err != nil {
			return fmt.Errorf("error creating self relation: %w", err)
		}
	}

	st.publishUpdate(Event{
		Action:     ObjectInserted,
		ObjectHash: objHash,
	})

	return nil
}

func (st *Store) putRelation(
	stream tilde.Digest,
	parent tilde.Digest,
	child tilde.Digest,
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
		return fmt.Errorf("could not prepare insert to objects table: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	_, err = stmt.Exec(
		stream.String(),
		parent.String(),
		child.String(),
	)
	if err != nil {
		return fmt.Errorf("could not insert to objects table: %w", err)
	}

	return nil
}

func (st *Store) GetStreamLeaves(
	streamRootHash tilde.Digest,
) ([]tilde.Digest, error) {
	st.tableLockObjects.Lock()
	defer st.tableLockObjects.Unlock()

	stmt, err := st.db.Prepare(`
		SELECT Parent
		FROM Relations
		WHERE
			RootHash=?
			AND Parent <> ''
			AND Parent NOT IN (
				SELECT DISTINCT Child
				FROM Relations
				WHERE
					RootHash=?
			)
	`)
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(streamRootHash.String(), streamRootHash.String())
	if err != nil {
		return nil, fmt.Errorf("could not query: %w", err)
	}
	defer rows.Close() // nolint: errcheck

	hashList := []tilde.Digest{}

	for rows.Next() {
		data := ""
		if err := rows.Scan(&data); err != nil {
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		hashList = append(hashList, tilde.Digest(data))
	}

	return hashList, nil
}

func (st *Store) GetRelations(
	parent tilde.Digest,
) ([]tilde.Digest, error) {
	st.tableLockObjects.Lock()
	defer st.tableLockObjects.Unlock()

	stmt, err := st.db.Prepare("SELECT Hash FROM Objects WHERE RootHash=?")
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(parent.String())
	if err != nil {
		return nil, fmt.Errorf("could not query: %w", err)
	}
	defer rows.Close() // nolint: errcheck

	hashList := []tilde.Digest{}

	for rows.Next() {
		data := ""
		if err := rows.Scan(&data); err != nil {
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		hashList = append(hashList, tilde.Digest(data))
	}

	istmt, err := st.db.Prepare(
		"UPDATE Objects SET LastAccessed=? WHERE RootHash=?",
	)
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}
	defer istmt.Close() // nolint: errcheck

	if _, err := istmt.Exec(
		time.Now().Unix(),
		parent.String(),
	); err != nil {
		return nil, fmt.Errorf("could not update last access: %w", err)
	}

	return hashList, nil
}

func (st *Store) ListHashes() ([]tilde.Digest, error) {
	st.tableLockObjects.Lock()
	defer st.tableLockObjects.Unlock()

	stmt, err := st.db.Prepare(
		"SELECT Hash FROM Objects WHERE Hash == RootHash",
	)
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("could not query: %w", err)
	}
	defer rows.Close() // nolint: errcheck

	hashList := []tilde.Digest{}

	for rows.Next() {
		data := ""
		if err := rows.Scan(&data); err != nil {
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		hashList = append(hashList, tilde.Digest(data))
	}

	return hashList, nil
}

func (st *Store) UpdateTTL(
	hash tilde.Digest,
	minutes int,
) error {
	st.tableLockObjects.Lock()
	defer st.tableLockObjects.Unlock()

	stmt, err := st.db.Prepare(`UPDATE Objects SET TTL=? WHERE RootHash=?`)
	if err != nil {
		return fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	if _, err := stmt.Exec(minutes, hash.String()); err != nil {
		return fmt.Errorf("could not update last access and ttl: %w", err)
	}

	return nil
}

func (st *Store) Remove(
	hash tilde.Digest,
) error {
	st.tableLockObjects.Lock()
	defer st.tableLockObjects.Unlock()

	stmt, err := st.db.Prepare(`
	DELETE FROM Objects
	WHERE Hash=?`)
	if err != nil {
		return fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	if _, err := stmt.Exec(
		hash.String(),
	); err != nil {
		return fmt.Errorf("could not delete object: %w", err)
	}

	st.publishUpdate(Event{
		Action:     ObjectRemoved,
		ObjectHash: hash,
	})

	return nil
}

func (st *Store) gc() error {
	st.tableLockObjects.Lock()
	defer st.tableLockObjects.Unlock()

	stmt, err := st.db.Prepare(`
	DELETE FROM Objects WHERE
		Hash NOT IN (
			SELECT Hash FROM Pins
		)
		AND TTL > 0
		AND datetime(LastAccessed + TTL, 'unixepoch') < datetime ('now');
	`)
	if err != nil {
		return fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	if _, err := stmt.Exec(); err != nil {
		return fmt.Errorf("could not gc delete objects: %w", err)
	}

	return nil
}

func (st *Store) Filter(
	filterOptions ...FilterOption,
) (object.ReadCloser, error) {
	st.tableLockObjects.Lock()
	defer st.tableLockObjects.Unlock()

	options := newFilterOptions(filterOptions...)

	where := "WHERE 1 "
	whereArgs := []interface{}{}

	if len(options.Filters.ObjectHashes) > 0 {
		qs := strings.Repeat(",?", len(options.Filters.ObjectHashes))[1:]
		where += "AND Hash IN (" + qs + ") "
		whereArgs = append(whereArgs, ahtoai(options.Filters.ObjectHashes)...)
	}

	if len(options.Filters.ContentTypes) > 0 {
		qs := strings.Repeat(",?", len(options.Filters.ContentTypes))[1:]
		where += "AND Type IN (" + qs + ") "
		whereArgs = append(whereArgs, astoai(options.Filters.ContentTypes)...)
	}

	if len(options.Filters.StreamHashes) > 0 {
		qs := strings.Repeat(",?", len(options.Filters.StreamHashes))[1:]
		where += "AND RootHash IN (" + qs + ") "
		whereArgs = append(whereArgs, ahtoai(options.Filters.StreamHashes)...)
	}

	if len(options.Filters.Owners) > 0 {
		qs := strings.Repeat(",?", len(options.Filters.Owners))[1:]
		where += "AND OwnerPublicKey IN (" + qs + ") "
		whereArgs = append(whereArgs, astoai(options.Filters.Owners)...)
	}

	where += fmt.Sprintf(
		"ORDER BY %s %s ",
		options.Filters.OrderBy,
		options.Filters.OrderDir,
	)

	if options.Filters.Limit != nil {
		where += fmt.Sprintf(
			"LIMIT %d ",
			*options.Filters.Limit,
		)
	}

	if options.Filters.Offset != nil {
		where += fmt.Sprintf(
			"OFFSET %d ",
			*options.Filters.Offset,
		)
	}

	// get the object
	// nolint: gosec
	stmt, err := st.db.Prepare("SELECT Hash FROM Objects " + where)
	if err != nil {
		return nil, fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(whereArgs...)
	if err != nil {
		return nil, fmt.Errorf("could not query: %w", err)
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
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		hashes = append(hashes, hash)
		hashesForUpdate = append(hashesForUpdate, hash)
	}

	if len(hashes) == 0 {
		return nil, objectstore.ErrNotFound
	}

	go func() {
		defer close(objectsChan)
		defer close(errorChan)
		for _, hash := range hashes {
			o, err := st.Get(tilde.Digest(hash))
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

func (st *Store) Pin(
	hash tilde.Digest,
) error {
	st.tableLockPins.Lock()
	defer st.tableLockPins.Unlock()

	stmt, err := st.db.Prepare(`
		INSERT OR IGNORE INTO Pins (Hash) VALUES (?)
	`)
	if err != nil {
		return fmt.Errorf("could not prepare insert to pins table, %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	_, err = stmt.Exec(
		hash,
	)
	if err != nil {
		return fmt.Errorf("could not insert to pins table, %w", err)
	}

	return nil
}

func (st *Store) GetPinned() ([]tilde.Digest, error) {
	st.tableLockPins.Lock()
	defer st.tableLockPins.Unlock()

	stmt, err := st.db.Prepare(`
		SELECT Hash FROM Pins
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

	hs := []tilde.Digest{}
	for rows.Next() {
		h := ""
		if err := rows.Scan(&h); err != nil {
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		if h != "" {
			hs = append(hs, tilde.Digest(h))
		}
	}

	return hs, nil
}

func (st *Store) IsPinned(hash tilde.Digest) (bool, error) {
	st.tableLockPins.Lock()
	defer st.tableLockPins.Unlock()

	stmt, err := st.db.Prepare(`
		SELECT Hash FROM Pins WHERE Hash = ?
	`)
	if err != nil {
		return false, fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(hash.String())
	if err != nil {
		return false, fmt.Errorf("could not query: %w", err)
	}
	defer rows.Close() // nolint: errcheck

	st.publishUpdate(Event{
		Action:     ObjectPinned,
		ObjectHash: hash,
	})

	if !rows.Next() {
		return false, nil
	}

	return true, nil
}

func (st *Store) RemovePin(
	hash tilde.Digest,
) error {
	st.tableLockPins.Lock()
	defer st.tableLockPins.Unlock()

	stmt, err := st.db.Prepare(`
		DELETE FROM Pins
		WHERE Hash=?
	`)
	if err != nil {
		return fmt.Errorf("could not prepare query, %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	st.publishUpdate(Event{
		Action:     ObjectUnpinned,
		ObjectHash: hash,
	})

	if _, err := stmt.Exec(
		hash.String(),
	); err != nil {
		return fmt.Errorf("could not delete object, %w", err)
	}

	return nil
}

func (st *Store) ListenForUpdates() (
	updates <-chan Event,
	cancel func(),
) {
	c := make(chan Event)
	st.listenersLock.Lock()
	defer st.listenersLock.Unlock()
	id := rand.String(8)
	st.listeners[id] = c
	f := func() {
		st.listenersLock.Lock()
		defer st.listenersLock.Unlock()
		delete(st.listeners, id)
	}
	return c, f
}

func (st *Store) publishUpdate(e Event) {
	st.listenersLock.Lock()
	defer st.listenersLock.Unlock()

	for _, l := range st.listeners {
		select {
		case l <- e:
		default:
		}
	}
}

func astoai(ah []string) []interface{} {
	as := make([]interface{}, len(ah))
	for i, h := range ah {
		as[i] = h
	}
	return as
}

func ahtoai(ah []tilde.Digest) []interface{} {
	as := make([]interface{}, len(ah))
	for i, h := range ah {
		as[i] = h.String()
	}
	return as
}

func (st *Store) PutKey(
	privateKey crypto.PrivateKey,
) error {
	st.tableLockKeys.Lock()
	defer st.tableLockKeys.Unlock()

	stmt, err := st.db.Prepare(`
		INSERT OR IGNORE INTO Keys (PublicKeyDigest, PrivateKey) VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("could not prepare insert to keys table, %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	data, _ := json.Marshal(privateKey) // nolint: errcheck // cannot error

	_, err = stmt.Exec(
		privateKey.PublicKey().Hash().String(),
		string(data),
	)
	if err != nil {
		return fmt.Errorf("could not insert to keys table, %w", err)
	}

	return nil
}

func (st *Store) GetKey(
	publicKeyDigest tilde.Digest,
) (*crypto.PrivateKey, error) {
	st.tableLockKeys.Lock()
	defer st.tableLockKeys.Unlock()

	stmt, err := st.db.Prepare(
		"SELECT PrivateKey FROM Keys WHERE PublicKeyDigest=?",
	)
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	row := stmt.QueryRow(publicKeyDigest.String())

	key := &crypto.PrivateKey{}
	data := []byte{}

	if err := row.Scan(&data); err != nil {
		return nil, errors.Merge(objectstore.ErrNotFound, err)
	}

	if err := json.Unmarshal(data, key); err != nil {
		return nil, fmt.Errorf("could not unmarshal key: %w", err)
	}

	return key, nil
}
