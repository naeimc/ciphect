package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"slices"
	"time"

	"golang.org/x/crypto/argon2"
)

type User struct {
	ID       string
	Created  time.Time
	Modified time.Time
	Username string
	HashType string
	Hash     []byte

	Exists bool
	Delete bool
}

func NewUser(username, password string) (User, error) {

	hash := NewArgon2IDHash(password)

	b, err := json.Marshal(hash)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:       ID(),
		Username: username,
		HashType: Argon2ID,
		Hash:     b,
	}, nil
}

func (user User) Validate(password string) error {
	switch user.HashType {
	case Argon2ID:
		var hash Argon2IDHash
		if err := json.Unmarshal(user.Hash, &hash); err != nil {
			return err
		}
		if !hash.Validate(password) {
			return ErrPasswordInvalid
		}
		return nil
	}
	return ErrUnknownHash
}

func (user User) QueryByUsername(ctx context.Context) (User, error) {
	return user.query(ctx, `SELECT id, created, modified, username, hash_type, hash FROM users WHERE username=$1`, user.Username)
}

func (user User) Query(ctx context.Context) (User, error) {
	return user.query(ctx, `SELECT id, created, modified, username, hash_type, hash FROM users WHERE id=$1`, user.ID)
}

func (user User) query(ctx context.Context, query string, a ...any) (usr User, err error) {
	row := DB.QueryRowContext(ctx, query, a...)
	usr = User{Exists: true}
	err = row.Scan(&usr.ID, &usr.Created, &usr.Modified, &usr.Username, &usr.HashType, &usr.Hash)
	return
}

func (user User) Update(tx *sql.Tx) error {
	if user.Exists {
		if user.Delete {
			return user.delete(tx)
		}
		return user.write(tx)
	}
	return user.create(tx)
}

func (user User) create(tx *sql.Tx) (err error) {
	_, err = tx.Exec(
		`INSERT INTO users (id, username, hash_type, hash) VALUES ($1, $2, $3, $4);`,
		user.ID,
		user.Username,
		user.HashType,
		user.Hash,
	)
	return err
}

func (user User) delete(tx *sql.Tx) (err error) {
	_, err = tx.Exec(
		`DELETE FROM users WHERE id=$1`,
		user.ID,
	)
	return err
}

func (user User) write(tx *sql.Tx) (err error) {
	_, err = tx.Exec(
		`UPDATE users SET username=$2, hash_type=$3, hash=$4 WHERE id=$1;`,
		user.ID,
		user.Username,
		user.HashType,
		user.Hash,
	)
	return err
}

const (
	Argon2ID        = "argon2id"
	Argon2IDTime    = 1
	Argon2IDMemory  = 64 * 1024
	Argon2IDThreads = 4
	Argon2IDLength  = 32
)

var (
	saltChannel chan []byte
)

func init() {
	saltChannel = make(chan []byte)
	go func() {
		for {
			b := make([]byte, Argon2IDLength)
			rand.Read(b)
			saltChannel <- b
		}
	}()
}

func Salt() []byte { return <-saltChannel }

type Argon2IDHash struct {
	Hash    []byte `json:"hash"`
	Salt    []byte `json:"salt"`
	Time    uint32 `json:"time"`
	Memory  uint32 `json:"memory"`
	Threads uint8  `json:"threads"`
	Length  uint32 `json:"length"`
}

func NewArgon2IDHash(password string) Argon2IDHash {
	salt := Salt()
	hash := argon2.IDKey([]byte(password), salt, Argon2IDTime, Argon2IDMemory, Argon2IDThreads, Argon2IDLength)
	return Argon2IDHash{hash, salt, Argon2IDTime, Argon2IDMemory, Argon2IDThreads, Argon2IDLength}
}

func (hash Argon2IDHash) Validate(password string) bool {
	b := argon2.IDKey([]byte(password), hash.Salt, hash.Time, hash.Memory, hash.Threads, hash.Length)
	return slices.Equal(hash.Hash, b)
}
