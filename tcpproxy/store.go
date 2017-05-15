package tcpproxy

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/boltdb/bolt"
)

var kBucket = []byte("portmapping")

func portToId(port int) []byte {
	id := fmt.Sprintf("%d", port)
	return []byte(id)
}

type Store struct {
	lock sync.Mutex
	db   *bolt.DB
}

func NewStore(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(kBucket)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) CleanAndUpdate(pms []*PortMappingInfo) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(kBucket)
		if err != nil {
			return err
		}

		b, err := tx.CreateBucketIfNotExists(kBucket)
		if err != nil {
			return err
		}

		for _, pm := range pms {
			data, err := json.Marshal(pm)
			if err != nil {
				return err
			}

			err = b.Put(portToId(pm.LocalPort), data)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (s *Store) AddPortMapping(pm *PortMappingInfo) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	data, err := json.Marshal(pm)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(kBucket)
		return b.Put(portToId(pm.LocalPort), data)
	})
}

func (s *Store) DeletePortMapping(localPort int) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(kBucket)
		return b.Delete(portToId(localPort))
	})
}

func (s *Store) GetAllPortMapping() ([]*PortMappingInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	result := make([]*PortMappingInfo, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			pm := &PortMappingInfo{}
			err := json.Unmarshal(v, pm)
			if err != nil {
				return err
			}
			result = append(result, pm)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
