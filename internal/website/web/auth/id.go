package auth

import (
	"crypto/rand"
	"fmt"
)

const IDLength = 16

var ids = make(chan string)

func init() {
	go func() {
		for {
			b := make([]byte, IDLength)
			rand.Read(b)
			ids <- fmt.Sprintf("%x", b)
		}
	}()
}

func ID() string { return <-ids }
