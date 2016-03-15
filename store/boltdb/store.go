package boltdb

import (
	"fmt"
	"github.com/Lupino/pusher"
	"github.com/boltdb/bolt"
)

// Store defined a bolt Storer interface
type Store struct {
	db     *bolt.DB
	path   string
	bucket string
	noSync bool
}

// New boltdb instance.
func New(config map[string]interface{}) (pusher.Storer, error) {
	var (
		path   string
		bucket string
		noSync bool
		ok     bool
		db     *bolt.DB
		err    error
	)
	path, ok = config["path"].(string)
	if !ok {
		return nil, fmt.Errorf("must specify path")
	}

	bucket, ok = config["bucket"].(string)
	if !ok {
		bucket = "pusher"
	}

	noSync, _ = config["nosync"].(bool)

	db, err = bolt.Open(path+"/pusher_store", 0600, nil)
	if err != nil {
		return nil, err
	}
	db.NoSync = noSync

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))

		return err
	})
	if err != nil {
		return nil, err
	}

	rv := Store{
		path:   path + "/pusher_store",
		bucket: bucket,
		db:     db,
		noSync: noSync,
	}
	return &rv, nil
}

// Set pusher into store
func (s Store) Set(p pusher.Pusher) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucket))
		err := b.Put([]byte(p.ID), p.Bytes())
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

// Get pusher from store
func (s Store) Get(p string) (pusher.Pusher, error) {
	var data []byte

	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucket))
		v := b.Get([]byte(p))
		size := len(v)
		data = make([]byte, size)
		copy(data[:size], v[:size])
		return nil
	})

	return pusher.NewPusher(data)

}

// Del pusher from store
func (s Store) Del(p string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucket))
		err := b.Delete([]byte(p))
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

// GetAll pusher from store
func (s Store) GetAll(from, size int) (uint64, []pusher.Pusher, error) {
	var hits []string
	var total uint64
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucket))
		c := b.Cursor()
		collector := newScorerSkipCollector(size, from)
		collector.Collect(c)
		total = collector.Total()
		ret := collector.Results()
		hits = make([]string, len(ret))
		copy(hits, ret)
		return nil
	})
	var pushers []pusher.Pusher
	for _, hit := range hits {
		p, _ := s.Get(hit)
		if p.ID == hit {
			pushers = append(pushers, p)
		}
	}
	return total, pushers, nil
}
