package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"
)

const KeyLength = 32

var keys = make(chan string)

func init() {
	go func() {
		for {
			b := make([]byte, KeyLength)
			rand.Read(b)
			keys <- fmt.Sprintf("%x", b)
		}
	}()
}

type Key struct {
	ID         string
	UserID     string
	Created    time.Time
	Name       string
	Session    bool
	Expires    bool
	Expiration time.Time

	Exists bool
	Delete bool
}

func NewNamedKey(user User, name string) Key {
	return newKey(user, name, false, false, time.Now().UTC())
}

func NewSessionKey(user User, expiration time.Time) Key {
	return newKey(user, "", true, true, expiration)
}

func newKey(user User, name string, session, expires bool, expiration time.Time) Key {
	return Key{
		ID:         <-keys,
		UserID:     user.ID,
		Name:       name,
		Session:    session,
		Expires:    expires,
		Expiration: expiration,
	}
}

func (key Key) Query(ctx context.Context) (k Key, err error) {
	row := DB.QueryRowContext(
		ctx,
		`SELECT id, user_id, created, name, session, expires, expiration FROM keys WHERE id=$1 AND session=$2`,
		key.ID,
		key.Session,
	)
	k = Key{Exists: true}
	if err = row.Scan(&k.ID, &k.UserID, &k.Created, &k.Name, &k.Session, &k.Expires, &k.Expiration); err != nil {
		return
	}
	return
}

func (key Key) Update(tx *sql.Tx) (err error) {
	if !key.Exists {
		return key.create(tx)
	}
	if key.Delete {
		return key.delete(tx)
	}
	return nil
}

func (key Key) create(tx *sql.Tx) (err error) {
	_, err = tx.Exec(
		`INSERT INTO keys (id, user_id, name, session, expires, expiration) VALUES ($1, $2, $3, $4, $5, $6);`,
		key.ID,
		key.UserID,
		key.Name,
		key.Session,
		key.Expires,
		key.Expiration,
	)
	return err

}

func (key Key) delete(tx *sql.Tx) (err error) {
	_, err = tx.Exec(
		`DELETE FROM keys WHERE id=$1;`,
		key.ID,
	)
	return err
}
