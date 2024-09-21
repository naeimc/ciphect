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
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/naeimc/ciphect/api"
	"github.com/naeimc/ciphect/exchange"
	"github.com/naeimc/ciphect/internal/website/logger"
	"github.com/naeimc/ciphect/internal/website/web/auth"
	"github.com/naeimc/ciphect/logging"
)

var Exchange = exchange.New()

var Group = new(sync.WaitGroup)

func init() {
	go Exchange.Start()
}

var ErrExchangeStopped = errors.New("exchange stopped")

func X(writer http.ResponseWriter, request *http.Request, session *auth.Session) (int, any) {

	if !session.Authenticated || session.Key.Name == "" {
		return http.StatusUnauthorized, http.StatusUnauthorized
	}

	information := map[string]string{
		"username": session.Username(),
		"key":      session.Key.Name,
	}

	name := "/" + session.Username() + "/" + session.Key.Name

	Group.Add(1)
	defer Group.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	endpoint, err := Exchange.OpenCtx(ctx, information, 1, name)
	if err != nil {
		switch err {
		case exchange.ErrEndpointExists:
			return http.StatusConflict, http.StatusConflict
		case ErrExchangeStopped:
			return http.StatusServiceUnavailable, http.StatusServiceUnavailable
		case context.Canceled:
			return http.StatusRequestTimeout, http.StatusRequestTimeout
		default:
			return http.StatusInternalServerError, http.StatusInternalServerError
		}
	}
	defer endpoint.Close(nil)

	connection, err := websocket.Accept(writer, request, nil)
	if err != nil {
		logger.Print(logging.Error, err)
		return http.StatusBadRequest, nil
	}
	defer connection.Close(websocket.StatusTryAgainLater, "")

	logger.Printf(logging.Information, "(CONNECTED) %s", endpoint.Name)
	defer logger.Printf(logging.Information, "(DISCONNECTED) %s", endpoint.Name)

	go writeToEndpoint(connection, endpoint)
	readFromEndpoint(connection, endpoint)

	return http.StatusOK, nil
}

func readFromEndpoint(connection *websocket.Conn, endpoint exchange.Endpoint) {
	for {

		messageType, b, err := connection.Read(context.Background())
		if err != nil {
			return
		}

		if messageType != websocket.MessageText {
			continue
		}

		go func(b []byte) {

			var packet api.Packet
			if err := json.Unmarshal(b, &packet); err != nil {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			lifespan := context.Background()
			if packet.Header.Expiration > 0 {
				var cancel context.CancelFunc
				lifespan, cancel = context.WithDeadline(ctx, packet.Header.Timestamp.Add(packet.Header.Expiration))
				defer cancel()
			}

			for _, to := range packet.Header.To {
				if err := endpoint.Send(lifespan, to, packet); err != nil {
					logger.Printf(logging.Warning, "(SEND) %d %s->%s %s", len(b), endpoint.Name, to, err)
				} else {
					logger.Printf(logging.Information, "(SEND) %d %s->%s", len(b), endpoint.Name, to)
				}
			}

		}(b)
	}
}

func writeToEndpoint(connection *websocket.Conn, endpoint exchange.Endpoint) {
	defer connection.CloseNow()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan func(), 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case f := <-c:
				f()
			}
		}
	}()

	for {
		header, data, closed, err := endpoint.Receive()
		if err != nil || closed {
			if err == ErrExchangeStopped {
				connection.Close(websocket.StatusGoingAway, err.Error())
			}
			return
		}

		b, err := json.Marshal(data)
		if err != nil {
			continue
		}

		go func() {
			c <- func() {
				if err := connection.Write(header.Lifespan, websocket.MessageText, b); err != nil {
					logger.Printf(logging.Warning, "(RECEIVE) %d %s<-%s %s", len(b), endpoint.Name, header.From, err)
				} else {
					logger.Printf(logging.Warning, "(RECEIVE) %d %s<-%s", len(b), endpoint.Name, header.From)
				}
			}
		}()
	}
}
