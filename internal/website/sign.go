/*
Ciphect, a personal data relay.
Copyright (C) 2024 Naeim Cragwell-Chaudhry

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package website

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/naeimc/ciphect/internal/website/html"
	"github.com/naeimc/ciphect/internal/website/logger"
	"github.com/naeimc/ciphect/internal/website/web"
	"github.com/naeimc/ciphect/internal/website/web/auth"
	"github.com/naeimc/ciphect/logging"
)

func SignUp(writer http.ResponseWriter, request *http.Request, session *auth.Session) (int, any) {

	if session.Authenticated {
		return http.StatusFound, web.Redirect{URL: "/"}
	}

	switch request.Method {
	case http.MethodPost:

		if err := request.ParseForm(); err != nil {
			logger.Print(logging.Error, err)
			break
		}

		username := request.FormValue("username")
		password1 := request.FormValue("password1")
		password2 := request.FormValue("password2")
		if password1 != password2 {
			break
		}

		if err := auth.SignUp(context.Background(), username, password1); err != nil {
			logger.Print(logging.Error, err)
			break
		}

		return http.StatusFound, web.Redirect{URL: "/"}
	}

	return http.StatusOK, web.HTMLTemplate{Template: html.SignUp, Data: nil}
}

func SignIn(writer http.ResponseWriter, request *http.Request, session *auth.Session) (int, any) {

	if session.Authenticated {
		return http.StatusFound, web.Redirect{URL: "/"}
	}

	var id string
	row := auth.DB.QueryRowContext(context.Background(), `SELECT id FROM users`)
	if err := row.Scan(&id); err == sql.ErrNoRows {
		return http.StatusFound, web.Redirect{URL: "/sign-up/"}
	}

	switch request.Method {
	case http.MethodPost:

		if err := request.ParseForm(); err != nil {
			logger.Print(logging.Error, err)
			break
		}

		username := request.FormValue("username")
		password := request.FormValue("password")

		cookie, err := auth.SignIn(context.Background(), username, password)
		if err != nil {
			logger.Print(logging.Error, err)
			break
		}

		session.Cookies = append(session.Cookies, cookie)
		return http.StatusFound, web.Redirect{URL: "/"}
	}

	return http.StatusOK, web.HTMLTemplate{Template: html.SignIn, Data: nil}
}

func SignOut(writer http.ResponseWriter, request *http.Request, session *auth.Session) (int, any) {
	if session.Authenticated {
		auth.SignOut(context.Background(), session.Key.ID)
	}

	session.Cookies = append(session.Cookies, auth.SignOut(context.Background(), session.Key.ID))
	return http.StatusFound, web.Redirect{URL: "/sign-in/"}
}
