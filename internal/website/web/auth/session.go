package auth

import (
	"net/http"
)

type Session struct {
	Authenticated bool
	Key           Key
	User          User
	Headers       map[string]string
	Cookies       []*http.Cookie
}

func (session Session) Username() string {
	if session.Authenticated {
		return session.User.Username
	}
	return "-"
}
