package internal

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

type Conn struct {
	*bolt.DB
}

func InitDB(path string, timeout time.Duration) (Conn, error) {
	db, err := bolt.Open(path, 0660, &bolt.Options{Timeout: timeout})
	if err != nil {
		return Conn{}, fmt.Errorf("open database: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("users"))
		return nil
	})

	if err != nil {
		return Conn{}, fmt.Errorf("create bucket: %v", err)
	}

	return Conn{db}, nil
}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func (c Conn) AddUser(user User) error {
	err := c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return fmt.Errorf(`bucket "users" does not exists`)
		}

		buf, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("marshal user: %w", err)
		}

		err = b.Put(itob(user.ID), buf)
		if err != nil {
			return fmt.Errorf("add user to db: %w", err)
		}

		return nil
	})

	return err
}

func (c Conn) GetUser(userID int64) (User, error) {
	var user User
	err := c.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return fmt.Errorf(`bucket "users" does not exists`)
		}

		key := itob(userID)

		buf := b.Get(key)
		if buf == nil {
			return ErrNotFound
		}

		err := json.Unmarshal(buf, &user)
		if err != nil {
			return fmt.Errorf("unmarshal user: %w", err)
		}

		return nil
	})

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (c Conn) AddDomain(userID int64, domain string) error {
	err := c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return fmt.Errorf(`bucket "users" does not exists`)
		}

		key := itob(userID)

		buf := b.Get(key)
		if buf == nil {
			return ErrNotFound
		}

		var user User
		err := json.Unmarshal(buf, &user)
		if err != nil {
			return fmt.Errorf("unmarshal user: %w", err)
		}

		user.ActiveDomains = append(user.ActiveDomains, domain)

		buf, err = json.Marshal(user)
		if err != nil {
			return fmt.Errorf("marshal user: %w", err)
		}
		b.Put(key, buf)

		return nil
	})

	return err
}

func (c Conn) DeleteDomain(userID int64, domain string) error {
	err := c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return fmt.Errorf(`bucket "users" does not exists`)
		}

		key := itob(userID)

		buf := b.Get(key)
		if buf == nil {
			return ErrNotFound
		}

		var user User
		err := json.Unmarshal(buf, &user)
		if err != nil {
			return fmt.Errorf("unmarshal user: %w", err)
		}

		flag := false
		for i, v := range user.ActiveDomains {
			if v == domain {
				flag = true
				user.ActiveDomains = append(user.ActiveDomains[:i], user.ActiveDomains[i+1:]...)
				break
			}
		}

		if !flag {
			return ErrNotFound
		}

		buf, err = json.Marshal(user)
		if err != nil {
			return fmt.Errorf("marshal user: %w", err)
		}
		b.Put(key, buf)

		return nil
	})

	return err
}
