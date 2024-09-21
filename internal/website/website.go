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

func Index(writer http.ResponseWriter, request *http.Request, session *auth.Session) (int, any) {
	if !session.Authenticated {
		return http.StatusFound, web.Redirect{URL: "/sign-in/"}
	}

	data := struct {
		Session   *auth.Session
		Keys      []keyEntry
		Endpoints []endpointEntry
	}{session, getKeys(session), getEndpoints(session)}

	return http.StatusOK, web.HTMLTemplate{Template: html.Index, Data: data}
}

func getKeys(session *auth.Session) (keys []keyEntry) {
	keys = make([]keyEntry, 0)

	rows, err := auth.DB.Query(`SELECT id, name FROM keys WHERE user_id=$1 AND session=$2`, session.User.ID, false)
	if err != nil {
		if err == sql.ErrNoRows {
			return
		}
		logger.Print(logging.Error, err)
		return
	}
	for rows.Next() {
		var key keyEntry
		if err := rows.Scan(&key.ID, &key.Name); err != nil {
			logger.Print(logging.Error, err)
			return
		}
		keys = append(keys, key)
	}

	return
}

type keyEntry struct {
	Name string
	ID   string
}

func getEndpoints(session *auth.Session) (endpoints []endpointEntry) {
	endpoints = make([]endpointEntry, 0)
	for _, endpoint := range Exchange.Endpoints {
		if endpoint.Information["username"] == session.Username() {
			endpoints = append(endpoints, endpointEntry{
				Name: endpoint.Name,
				Key:  endpoint.Information["key"]})
		}
	}
	return
}

type endpointEntry struct {
	Name string
	Key  string
}

func KeyCreate(writer http.ResponseWriter, request *http.Request, session *auth.Session) (int, any) {
	if !session.Authenticated {
		return http.StatusFound, web.Redirect{URL: "/sign-in/"}
	}

	if err := request.ParseForm(); err != nil {
		logger.Print(logging.Error, err)
	}

	name := request.FormValue("name")

	if name != "" {
		key := auth.NewNamedKey(session.User, name)
		if err := auth.Transact(context.Background(), key.Update); err != nil {
			logger.Print(logging.Error, err)
		}
	}

	return http.StatusFound, web.Redirect{URL: "/"}
}

func KeyDelete(writer http.ResponseWriter, request *http.Request, session *auth.Session) (int, any) {
	if !session.Authenticated {
		return http.StatusFound, web.Redirect{URL: "/sign-in/"}
	}

	if err := request.ParseForm(); err != nil {
		logger.Print(logging.Error, err)
	}

	id := request.FormValue("id")

	if id != "" {
		key, err := auth.Key{ID: id}.Query(context.Background())
		if err != nil {
			logger.Print(logging.Error, err)
		} else {
			key.Delete = true
			if err := auth.Transact(context.Background(), key.Update); err != nil {
				logger.Print(logging.Error, err)
			}
		}
	}

	return http.StatusFound, web.Redirect{URL: "/"}
}
