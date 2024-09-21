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

package cli

import (
	"runtime"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"

	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/naeimc/ciphect/internal/website"
	"github.com/naeimc/ciphect/internal/website/logger"
	"github.com/naeimc/ciphect/internal/website/web"
	"github.com/naeimc/ciphect/internal/website/web/auth"
	"github.com/naeimc/ciphect/logging"
)

func run() (code int) {

	var (
		address     = os.Getenv("CIPHECT_ADDRESS")
		certificate = os.Getenv("CIPHECT_CERTIFICATE_FILE")
		key         = os.Getenv("CIPHECT_KEY_FILE")
		database    = os.Getenv("CIPHECT_DATABASE")
		cookie      = os.Getenv("CIPHECT_COOKIE")
	)

	logger.Logger = logging.New(logging.Information, runtime.NumCPU(), logging.String{Writer: os.Stderr})
	defer logger.Logger.Close()

	db, err := sql.Open("pgx", database)
	if err != nil {
		logger.Print(logging.Fatal, err)
		code = 1
		return
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {})

	web.HandleFunction(handler, "/", website.Index)
	web.HandleFunction(handler, "/x/", website.X)
	web.HandleFunction(handler, "POST /key/create/", website.KeyCreate)
	web.HandleFunction(handler, "POST /key/delete/", website.KeyDelete)
	web.HandleFunction(handler, "/sign-up/", website.SignUp)
	web.HandleFunction(handler, "/sign-in/", website.SignIn)
	web.HandleFunction(handler, "/sign-out/", website.SignOut)

	server := &web.Server{
		Address:     address,
		Handler:     handler,
		Certificate: certificate,
		Key:         key,

		StopSignals: []os.Signal{os.Interrupt, syscall.SIGTERM},
		StopTimeout: time.Second * 10,

		OnShutdown: []func(){func() { website.Exchange.Stop(website.ErrExchangeStopped) }},
	}

	auth.DB = db
	auth.CookieName = cookie
	auth.CookieSecure = server.UseTLS()

	if err := server.ListenAndServe(); err != nil {
		code = 1
	}

	website.Group.Wait()

	return
}
