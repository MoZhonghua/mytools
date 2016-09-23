package tcpmux

import (
	"database/sql"
	"encoding/json"
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

	_, err = db.Exec(`create table if not exists target 
	(id text primary key, data text)`)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) CleanAndUpdate(pms []*TargetInfo) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	err = func() error {
		_, err := tx.Exec("delete from target")
		if err != nil {
			return err
		}
		for _, pm := range pms {
			data, err := json.Marshal(pm)
			if err != nil {
				return err
			}

			_, err = tx.Exec("insert into target(id, data) values (?, ?)",
				pm.Id, data)
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

func (s *Store) AddTarget(pm *TargetInfo) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	data, err := json.Marshal(pm)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("insert into target(id, data) values (?, ?)", pm.Id, data)
	return err
}

func (s *Store) DeleteTarget(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, err := s.db.Exec("delete from target where id = ?", id)
	return err
}

func (s *Store) GetAllTarget() ([]*TargetInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	rows, err := s.db.Query("select data from target")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*TargetInfo, 0)
	for rows.Next() {
		var data string
		err := rows.Scan(&data)
		if err != nil {
			return nil, err
		}

		pm := &TargetInfo{}
		err = json.Unmarshal([]byte(data), pm)
		if err != nil {
			return nil, err
		}
		result = append(result, pm)
	}

	return result, nil
}
