package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrPasswordInvalid = errors.New("password invalid")
	ErrUnknownHash     = errors.New("unknown hash")
	ErrKeyNotFound     = errors.New("key not found")
	ErrKeyExpired      = errors.New("key expired")
)

var (
	CookieName   string
	CookieSecure bool
)

func SignUp(ctx context.Context, username, password string) (err error) {

	user, err := NewUser(username, password)
	if err != nil {
		return err
	}

	return Transact(ctx, user.Update)
}

func SignIn(ctx context.Context, username, password string) (cookie *http.Cookie, err error) {

	user, err := (User{Username: username}).QueryByUsername(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if err := user.Validate(password); err != nil {
		return nil, err
	}

	key := NewSessionKey(user, time.Now().UTC().Add(time.Hour*4))
	if err := Transact(ctx, key.Update); err != nil {
		return nil, err
	}

	cookie = &http.Cookie{
		Name:     CookieName,
		Value:    key.ID,
		Path:     "/",
		MaxAge:   int(time.Until(key.Expiration).Seconds()),
		Secure:   CookieSecure,
		HttpOnly: true,
	}

	return cookie, nil
}

func SignOut(ctx context.Context, keyID string) *http.Cookie {

	key, err := (Key{ID: keyID}).Query(ctx)
	if err == nil {
		key.Delete = true
		Transact(ctx, key.Update)
	}

	cookie := &http.Cookie{
		Name:     CookieName,
		Path:     "/",
		MaxAge:   -1,
		Secure:   CookieSecure,
		HttpOnly: true,
	}

	return cookie
}

func Authenticate(ctx context.Context, request *http.Request) (session *Session, err error) {

	session = &Session{}

	keyID, _, ok := request.BasicAuth()
	if !ok {
		cookie, err := request.Cookie(CookieName)
		if err != nil {
			if err == http.ErrNoCookie {
				err = nil
			}
			return session, err
		}
		keyID = cookie.Value
	}

	key, err := Key{ID: keyID, Session: !ok}.Query(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return
	}

	if key.Expires && time.Now().After(key.Expiration) {
		key.Delete = true
		Transact(ctx, key.Update)
		return
	}

	user, err := User{ID: key.UserID}.Query(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return
	}

	session.Authenticated = true
	session.Key = key
	session.User = user
	return
}
