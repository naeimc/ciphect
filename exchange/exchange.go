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

package exchange

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrEndpointDoesNotExist = errors.New("endpoint does not exist")
	ErrEndpointExists       = errors.New("endpoint exists")
)

type Exchange struct {
	Endpoints map[string]Endpoint

	queue chan work
	group *sync.WaitGroup
	stop  error
}

func New() *Exchange {
	return &Exchange{Endpoints: make(map[string]Endpoint), queue: make(chan work), group: new(sync.WaitGroup)}
}

func (exchange *Exchange) Start() {
	for work := range exchange.queue {
		if exchange.stop != nil && !work.close {
			work.err <- exchange.stop
		} else {
			work.err <- work.function()
		}
	}
}

func (exchange *Exchange) Stop(reason error) (err error) {
	return exchange.StopCtx(context.Background(), reason)
}

func (exchange *Exchange) StopCtx(ctx context.Context, reason error) (err error) {
	exchange.do(ctx, false, func() error {
		exchange.stop = reason
		return nil
	})
	for _, endpoint := range exchange.Endpoints {
		endpoint.CloseCtx(ctx, reason)
	}
	exchange.group.Wait()
	return nil
}

func (exchange *Exchange) Open(information map[string]string, size int, name string) (endpoint Endpoint, err error) {
	return exchange.OpenCtx(context.Background(), information, size, name)
}

func (exchange *Exchange) OpenCtx(ctx context.Context, information map[string]string, size int, name string) (endpoint Endpoint, err error) {
	return exchange.open(ctx, information, size, name, 0)
}

func (exchange *Exchange) OpenWildcard(information map[string]string, size int, format string, length int) (endpoint Endpoint, err error) {
	return exchange.OpenWildcardCtx(context.Background(), information, size, format, length)
}

func (exchange *Exchange) OpenWildcardCtx(ctx context.Context, information map[string]string, size int, format string, length int) (endpoint Endpoint, err error) {
	return exchange.open(ctx, information, size, format, length)
}

func (exchange *Exchange) open(ctx context.Context, information map[string]string, size int, format string, length int) (endpoint Endpoint, err error) {
	if information == nil {
		information = make(map[string]string)
	}

	err = exchange.do(ctx, false, func() error {
		name, err := exchange.wildcard(ctx, format, length)
		if err != nil {
			return err
		}
		if _, found := exchange.Endpoints[name]; found {
			return ErrEndpointExists
		}
		endpoint = Endpoint{
			Name:        name,
			Information: information,
			C:           make(chan Message, size),
			exchange:    exchange,
		}
		exchange.Endpoints[name] = endpoint
		return nil
	})

	if err != nil {
		return
	}
	return
}

func (exchange *Exchange) wildcard(ctx context.Context, format string, length int) (string, error) {
	if length == 0 || !strings.Contains(format, "*") {
		return format, nil
	}

	c := make(chan string, 1)

	go func() {
		b := make([]byte, length)
		rand.Read(b)
		name := strings.ReplaceAll(format, "*", fmt.Sprintf("%x", b))
		if _, found := exchange.Endpoints[name]; !found {
			c <- name
			return
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case name := <-c:
			return name, nil
		}
	}
}

func (exchange *Exchange) Close(name string, reason error) error {
	return exchange.CloseCtx(context.Background(), name, reason)
}

func (exchange *Exchange) CloseCtx(ctx context.Context, name string, reason error) error {
	return exchange.do(ctx, true, func() error {
		endpoint, found := exchange.Endpoints[name]
		if found {
			delete(exchange.Endpoints, name)
			exchange.group.Add(1)
			go func() {
				defer exchange.group.Done()
				endpoint.C <- Message{Error: reason, Closed: true}
			}()
		}
		return nil
	})
}

func (exchange *Exchange) Send(lifespan context.Context, to, from string, data any) error {
	return exchange.SendCtx(context.Background(), lifespan, to, from, data)
}

func (exchange *Exchange) SendCtx(ctx context.Context, lifespan context.Context, to, from string, data any) error {
	return exchange.do(ctx, false, func() error {

		endpoint, found := exchange.Endpoints[to]
		if !found {
			return ErrEndpointDoesNotExist
		}
		exchange.group.Add(1)
		go func() {
			defer exchange.group.Done()
			endpoint.C <- Message{Header: Header{
				Lifespan: lifespan,
				To:       to,
				From:     from,
			}, Data: data}
		}()
		return nil
	})
}

func (exchange *Exchange) do(ctx context.Context, close bool, function func() error) error {
	err := make(chan error, 1)
	exchange.queue <- work{function, err, close}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-err:
		return err
	}
}

type Endpoint struct {
	Name        string
	Information map[string]string
	C           chan Message

	exchange *Exchange
}

func (endpoint Endpoint) Send(lifespan context.Context, to string, data any) error {
	return endpoint.SendCtx(context.Background(), lifespan, to, data)
}

func (endpoint Endpoint) SendCtx(ctx context.Context, lifespan context.Context, to string, data any) error {
	return endpoint.exchange.SendCtx(ctx, lifespan, to, endpoint.Name, data)
}

func (endpoint Endpoint) Receive() (header Header, a any, closed bool, err error) {
	return endpoint.ReceiveCtx(context.Background())
}

func (endpoint Endpoint) ReceiveCtx(ctx context.Context) (header Header, a any, closed bool, err error) {
	select {
	case <-ctx.Done():
		return Header{}, nil, false, ctx.Err()
	case message := <-endpoint.C:
		return message.Header, message.Data, message.Closed, message.Error
	}
}

func (endpoint Endpoint) Close(reason error) error {
	return endpoint.exchange.Close(endpoint.Name, reason)
}

func (endpoint Endpoint) CloseCtx(ctx context.Context, reason error) error {
	return endpoint.exchange.CloseCtx(ctx, endpoint.Name, reason)
}

type Message struct {
	Header Header
	Data   any
	Error  error
	Closed bool
}

type Header struct {
	To, From string
	Lifespan context.Context
}

type work struct {
	function func() error
	err      chan error
	close    bool
}
