package sqlobjectstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	// required for sqlite3
	_ "github.com/mattn/go-sqlite3"

	"nimona.io/internal/rand"
	"nimona.io/pkg/chore"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/migration"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
)

// nolint: lll
var migrations = []string{
	`CREATE TABLE IF NOT EXISTS Objects (CID TEXT NOT NULL PRIMARY KEY);`,
	`ALTER TABLE Objects ADD Type TEXT;`,
	`ALTER TABLE Objects ADD Body TEXT;`,
	`ALTER TABLE Objects ADD RootCID TEXT;`,
	`ALTER TABLE Objects ADD TTL INT;`,
	`ALTER TABLE Objects ADD Created INT;`,
	`ALTER TABLE Objects ADD LastAccessed INT;`,
	`ALTER TABLE Objects ADD SignerPublicKey TEXT;`,
	`ALTER TABLE Objects ADD AuthorPublicKey TEXT;`,
	`ALTER TABLE Objects RENAME AuthorPublicKey TO OwnerPublicKey;`,
	`ALTER TABLE Objects RENAME SignerPublicKey TO _DeprecatedSignerPublicKey;`,
	`CREATE INDEX Created_idx ON Objects(Created);`,
	`CREATE INDEX TTL_LastAccessed_idx ON Objects(TTL, LastAccessed);`,
	`CREATE INDEX Type_RootCID_OwnerPublicKey_idx ON Objects(Type, RootCID, OwnerPublicKey);`,
	`CREATE INDEX RootCID_idx ON Objects(RootCID);`,
	`CREATE INDEX RootCID_TTL_idx ON Objects(RootCID, TTL);`,
	`CREATE INDEX CID_LastAccessed_idx ON Objects(CID, LastAccessed);`,
	`CREATE TABLE IF NOT EXISTS Relations (Parent TEXT NOT NULL, Child TEXT NOT NULL, PRIMARY KEY (Parent, Child));`,
	`ALTER TABLE Relations ADD RootCID TEXT;`,
	`CREATE INDEX Relations_RootCID_idx ON Relations(RootCID);`,
	`ALTER TABLE Objects ADD MetadataDatetime INT DEFAULT 0;`,
	`CREATE TABLE IF NOT EXISTS Pins (CID TEXT NOT NULL PRIMARY KEY);`,
}

var defaultTTL = time.Hour * 24 * 7

type (
	Store struct {
		db            *sql.DB
		listeners     map[string]chan Event
		listenersLock sync.RWMutex
	}
	EventAction string
	Event       struct {
		Action    EventAction
		ObjectCID chore.CID
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
		db:            db,
		listeners:     map[string]chan Event{},
		listenersLock: sync.RWMutex{},
	}

	// run migrations
	if err := migration.Up(db, migrations...); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
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
	cid chore.CID,
) (*object.Object, error) {
	// get the object
	stmt, err := st.db.Prepare("SELECT Body FROM Objects WHERE CID=?")
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	row := stmt.QueryRow(cid.String())

	obj := &object.Object{}
	data := []byte{}

	if err := row.Scan(&data); err != nil {
		return nil, errors.Merge(objectstore.ErrNotFound, err)
	}

	if err := json.Unmarshal(data, obj); err != nil {
		return nil, fmt.Errorf("could not unmarshal data: %w", err)
	}

	// update the last accessed column
	istmt, err := st.db.Prepare(
		"UPDATE Objects SET LastAccessed=? WHERE CID=?")
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}
	defer istmt.Close() // nolint: errcheck

	if _, err := istmt.Exec(
		time.Now().Unix(),
		cid.String(),
	); err != nil {
		return nil, fmt.Errorf("could not update last access: %w", err)
	}

	return obj, nil
}

func (st *Store) GetByStream(
	streamRootCID chore.CID,
) (object.ReadCloser, error) {
	return st.Filter(
		FilterByStreamCID(streamRootCID),
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
	// TODO(geoah) why replace?
	stmt, err := st.db.Prepare(`
	REPLACE INTO Objects (
		CID,
		Type,
		RootCID,
		OwnerPublicKey,
		Body,
		Created,
		LastAccessed,
		TTL,
		MetadataDatetime
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?
	) ON CONFLICT (CID) DO UPDATE SET
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

	objCID := obj.CID()
	objectType := obj.Type
	objectCID := objCID.String()
	streamCID := obj.Metadata.Stream.String()
	// TODO support multiple owners
	ownerPublicKey := ""
	if !obj.Metadata.Owner.IsEmpty() {
		ownerPublicKey = obj.Metadata.Owner.String()
	}

	// if the object doesn't belong to a stream, we need to set the stream
	// to the object's cid.
	// This should allow queries to consider the root object part of the stream.
	if streamCID == "" {
		streamCID = objectCID
	}

	un := 0
	dt, err := time.Parse(
		time.RFC3339,
		obj.Metadata.Datetime,
	)
	if err == nil {
		un = int(dt.Unix())
	}

	_, err = stmt.Exec(
		// VALUES
		objectCID,
		objectType,
		streamCID,
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
				err := st.putRelation(chore.CID(streamCID), objCID, p)
				if err != nil {
					return fmt.Errorf("could not create relation: %w", err)
				}
			}
		}
	}

	if streamCID == objectCID {
		err := st.putRelation(chore.CID(streamCID), objCID, "")
		if err != nil {
			return fmt.Errorf("error creating self relation: %w", err)
		}
	}

	st.publishUpdate(Event{
		Action:    ObjectInserted,
		ObjectCID: objCID,
	})

	return nil
}

func (st *Store) putRelation(
	stream chore.CID,
	parent chore.CID,
	child chore.CID,
) error {
	stmt, err := st.db.Prepare(`
		INSERT OR IGNORE INTO Relations (
			RootCID,
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
	streamRootCID chore.CID,
) ([]chore.CID, error) {
	stmt, err := st.db.Prepare(`
		SELECT Parent
		FROM Relations
		WHERE
			RootCID=?
			AND Parent <> ''
			AND Parent NOT IN (
				SELECT DISTINCT Child
				FROM Relations
				WHERE
					RootCID=?
			)
	`)
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(streamRootCID.String(), streamRootCID.String())
	if err != nil {
		return nil, fmt.Errorf("could not query: %w", err)
	}
	defer rows.Close() // nolint: errcheck

	cidList := []chore.CID{}

	for rows.Next() {
		data := ""
		if err := rows.Scan(&data); err != nil {
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		cidList = append(cidList, chore.CID(data))
	}

	return cidList, nil
}

func (st *Store) GetRelations(
	parent chore.CID,
) ([]chore.CID, error) {
	stmt, err := st.db.Prepare("SELECT CID FROM Objects WHERE RootCID=?")
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(parent.String())
	if err != nil {
		return nil, fmt.Errorf("could not query: %w", err)
	}
	defer rows.Close() // nolint: errcheck

	cidList := []chore.CID{}

	for rows.Next() {
		data := ""
		if err := rows.Scan(&data); err != nil {
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		cidList = append(cidList, chore.CID(data))
	}

	istmt, err := st.db.Prepare(
		"UPDATE Objects SET LastAccessed=? WHERE RootCID=?",
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

	return cidList, nil
}

func (st *Store) ListCIDs() ([]chore.CID, error) {
	stmt, err := st.db.Prepare(
		"SELECT CID FROM Objects WHERE CID == RootCID",
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

	cidList := []chore.CID{}

	for rows.Next() {
		data := ""
		if err := rows.Scan(&data); err != nil {
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		cidList = append(cidList, chore.CID(data))
	}

	return cidList, nil
}

func (st *Store) UpdateTTL(
	cid chore.CID,
	minutes int,
) error {
	stmt, err := st.db.Prepare(`UPDATE Objects SET TTL=? WHERE RootCID=?`)
	if err != nil {
		return fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	if _, err := stmt.Exec(minutes, cid.String()); err != nil {
		return fmt.Errorf("could not update last access and ttl: %w", err)
	}

	return nil
}

func (st *Store) Remove(
	cid chore.CID,
) error {
	stmt, err := st.db.Prepare(`
	DELETE FROM Objects
	WHERE CID=?`)
	if err != nil {
		return fmt.Errorf("could not prepare query: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	if _, err := stmt.Exec(
		cid.String(),
	); err != nil {
		return fmt.Errorf("could not delete object: %w", err)
	}

	st.publishUpdate(Event{
		Action:    ObjectRemoved,
		ObjectCID: cid,
	})

	return nil
}

func (st *Store) gc() error {
	stmt, err := st.db.Prepare(`
	DELETE FROM Objects WHERE
		CID NOT IN (
			SELECT CID FROM Pins
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
	options := newFilterOptions(filterOptions...)

	where := "WHERE 1 "
	whereArgs := []interface{}{}

	if len(options.Filters.ObjectCIDs) > 0 {
		qs := strings.Repeat(",?", len(options.Filters.ObjectCIDs))[1:]
		where += "AND CID IN (" + qs + ") "
		whereArgs = append(whereArgs, ahtoai(options.Filters.ObjectCIDs)...)
	}

	if len(options.Filters.ContentTypes) > 0 {
		qs := strings.Repeat(",?", len(options.Filters.ContentTypes))[1:]
		where += "AND Type IN (" + qs + ") "
		whereArgs = append(whereArgs, astoai(options.Filters.ContentTypes)...)
	}

	if len(options.Filters.StreamCIDs) > 0 {
		qs := strings.Repeat(",?", len(options.Filters.StreamCIDs))[1:]
		where += "AND RootCID IN (" + qs + ") "
		whereArgs = append(whereArgs, ahtoai(options.Filters.StreamCIDs)...)
	}

	if len(options.Filters.Owners) > 0 {
		qs := strings.Repeat(",?", len(options.Filters.Owners))[1:]
		where += "AND OwnerPublicKey IN (" + qs + ") "
		whereArgs = append(whereArgs, aktoai(options.Filters.Owners)...)
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
	stmt, err := st.db.Prepare("SELECT CID FROM Objects " + where)
	if err != nil {
		return nil, fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(whereArgs...)
	if err != nil {
		return nil, fmt.Errorf("could not query: %w", err)
	}
	defer rows.Close() // nolint: errcheck

	cids := []string{}
	cidsForUpdate := []interface{}{}

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
		cid := ""
		if err := rows.Scan(&cid); err != nil {
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		cids = append(cids, cid)
		cidsForUpdate = append(cidsForUpdate, cid)
	}

	if len(cids) == 0 {
		return nil, objectstore.ErrNotFound
	}

	// update the last accessed column
	updateQs := strings.Repeat(",?", len(cids))[1:]
	istmt, err := st.db.Prepare(
		"UPDATE Objects SET LastAccessed = ? " +
			"WHERE CID IN (" + updateQs + ")",
	)
	if err != nil {
		return nil, err
	}
	defer istmt.Close() // nolint: errcheck

	if _, err := istmt.Exec(
		append([]interface{}{time.Now().Unix()}, cidsForUpdate...)...,
	); err != nil {
		return nil, err
	}

	go func() {
		defer close(objectsChan)
		defer close(errorChan)
		for _, cid := range cids {
			o, err := st.Get(chore.CID(cid))
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
	cid chore.CID,
) error {
	stmt, err := st.db.Prepare(`
		INSERT OR IGNORE INTO Pins (CID) VALUES (?)
	`)
	if err != nil {
		return fmt.Errorf("could not prepare insert to pins table, %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	_, err = stmt.Exec(
		cid,
	)
	if err != nil {
		return fmt.Errorf("could not insert to pins table, %w", err)
	}

	return nil
}

func (st *Store) GetPinned() ([]chore.CID, error) {
	stmt, err := st.db.Prepare(`
		SELECT CID FROM Pins
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

	hs := []chore.CID{}
	for rows.Next() {
		h := ""
		if err := rows.Scan(&h); err != nil {
			return nil, errors.Merge(objectstore.ErrNotFound, err)
		}
		if h != "" {
			hs = append(hs, chore.CID(h))
		}
	}

	return hs, nil
}

func (st *Store) IsPinned(cid chore.CID) (bool, error) {
	stmt, err := st.db.Prepare(`
		SELECT CID FROM Pins WHERE CID = ?
	`)
	if err != nil {
		return false, fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	rows, err := stmt.Query(cid.String())
	if err != nil {
		return false, fmt.Errorf("could not query: %w", err)
	}
	defer rows.Close() // nolint: errcheck

	st.publishUpdate(Event{
		Action:    ObjectPinned,
		ObjectCID: cid,
	})

	if !rows.Next() {
		return false, nil
	}

	return true, nil
}

func (st *Store) RemovePin(
	cid chore.CID,
) error {
	stmt, err := st.db.Prepare(`
		DELETE FROM Pins
		WHERE CID=?
	`)
	if err != nil {
		return fmt.Errorf("could not prepare query, %w", err)
	}
	defer stmt.Close() // nolint: errcheck

	st.publishUpdate(Event{
		Action:    ObjectUnpinned,
		ObjectCID: cid,
	})

	if _, err := stmt.Exec(
		cid.String(),
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
	st.listenersLock.RLock()
	defer st.listenersLock.RUnlock()
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

func ahtoai(ah []chore.CID) []interface{} {
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
