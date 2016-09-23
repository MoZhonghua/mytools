package tcpproxy

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	lock sync.Mutex
	db   *sql.DB
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`create table if not exists portmapping 
	(id text primary key, data text)`)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) CleanAndUpdate(pms []*PortMappingInfo) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	err = func() error {
		_, err := tx.Exec("delete from portmapping")
		if err != nil {
			return err
		}
		for _, pm := range pms {
			data, err := json.Marshal(pm)
			if err != nil {
				return err
			}

			id := fmt.Sprintf("%d", pm.LocalPort)
			_, err = tx.Exec("insert into portmapping(id, data) values (?, ?)",
				id, data)
			return err
		}
		return nil
	}()

	if err == nil {
		return tx.Commit()
	} else {
		tx.Rollback()
		return err
	}
}

func (s *Store) AddPortMapping(pm *PortMappingInfo) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	data, err := json.Marshal(pm)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%d", pm.LocalPort)

	_, err = s.db.Exec("insert into portmapping(id, data) values (?, ?)", id, data)
	return err
}

func (s *Store) DeletePortMapping(localPort int) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	id := fmt.Sprintf("%d", localPort)
	_, err := s.db.Exec("delete from portmapping where id = ?", id)
	return err
}

func (s *Store) GetAllPortMapping() ([]*PortMappingInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	rows, err := s.db.Query("select data from portmapping")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*PortMappingInfo, 0)
	for rows.Next() {
		var data string
		err := rows.Scan(&data)
		if err != nil {
			return nil, err
		}

		pm := &PortMappingInfo{}
		err = json.Unmarshal([]byte(data), pm)
		if err != nil {
			return nil, err
		}
		result = append(result, pm)
	}

	return result, nil
}
