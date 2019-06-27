package backlog

import (
	"encoding/json"
	"time"

	"github.com/asdine/storm"
	query "github.com/asdine/storm/q"

	"nimona.io/internal/errors"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

type (
	// Bolt is a BoltDB backed Backlog
	Bolt struct {
		storm *storm.DB
	}
	// key is our composite key for keeping track of our backlog
	key struct {
		ObjectHash string
		KeyHash    string
	}
	// item holds the object we want to send, and some additional metadata
	item struct {
		ID     int `storm:"id,increment"`
		Key    key `storm:"unique"`
		Object []byte
		Pushed time.Time `storm:"index"`

		// TODO(geoah) figure out how to select using only one part of the
		// composite key
		ObjectHash string `storm:"index"`
		KeyHash    string `storm:"index"`
	}
)

// NewBolt constructs a new boltdb backed backlog, given a storm instance.
func NewBolt(st *storm.DB) (Backlog, error) {
	bl := &Bolt{
		storm: st,
	}
	go bl.gc()
	return bl, nil
}

// Push an object to the backlog with one or more recipients
func (bl *Bolt) Push(o *object.Object, ks ...*crypto.PublicKey) error {
	b, err := json.Marshal(o.ToMap())
	if err != nil {
		return err
	}

	h := o.HashBase58()
	n := time.Now()
	for _, k := range ks {
		err := bl.storm.Save(&item{
			Key: key{
				ObjectHash: h,
				KeyHash:    k.Fingerprint().String(),
			},
			Object: b,
			Pushed: n,
			// TODO(geoah) remove once we figured out how to select based on
			// only one part of the composite key
			ObjectHash: h,
			KeyHash:    k.Fingerprint().String(),
		})
		if err != nil {
			switch err {
			case storm.ErrAlreadyExists:
				return errors.Wrap(err, ErrAlreadyExists)
			default:
				return err
			}
		}
	}
	return nil
}

// Pop an object from the backlog for a specific recipient
func (bl *Bolt) Pop(k *crypto.PublicKey) (*object.Object, AckFunc, error) {
	item := &item{}
	q := bl.storm.Select(
		query.Eq("KeyHash", k.Fingerprint().String()),
	).
		OrderBy("Pushed")

	if err := q.First(item); err != nil {
		switch err {
		case storm.ErrNotFound:
			return nil, nil, errors.Wrap(err, ErrNoMoreObjects)
		default:
			return nil, nil, err
		}
	}

	// TODO there is a race condition if Pop() gets called more than once
	// at the same time. We should be locking the item somehow when returning.
	// In SQL this would be something like `SELECT FOR UPDATE` probably.
	if err := bl.storm.DeleteStruct(item); err != nil {
		return nil, nil, err
	}

	o := &object.Object{}
	m := map[string]interface{}{}
	if err := json.Unmarshal(item.Object, &m); err != nil {
		return nil, nil, err
	}

	if err := o.FromMap(m); err != nil {
		return nil, nil, err
	}

	return o, nil, nil
}

// garbage collection to clean up bolt db from old objects
func (bl *Bolt) gc() {
	// TODO implement garbage collection
	// rmOlderThan := time.Second * 60 * 60 * 24
	// ticker := time.NewTicker(60 * time.Second)
	// for {
	// 	select {
	// 	case <-ticker.C:
	// 		q := bl.storm.Select(
	// 			query.Gt("Pushed", time.Now().Add(-rmOlderThan)),
	// 		).OrderBy("Pushed")
	// 		if err := q.Delete(item{}); err != nil {
	// 			continue
	// 		}
	// 	}
	// }
}
