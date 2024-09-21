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
	"fmt"
	"testing"
	"time"
)

func TestExchange(t *testing.T) {

	exchange := New()
	exchangeReady := make(chan any)

	var receiver string
	receiverReady := make(chan any)
	receiverDone := make(chan any)
	receiverError := make(chan any)

	senderDone := make(chan any)

	message := "Hello, World!"
	errStopped := fmt.Errorf("exchange stopped")

	t.Run("Start|Stop", func(t *testing.T) {
		t.Parallel()
		t.Log("starting exchange")
		go exchange.Start()
		close(exchangeReady)
		<-receiverDone
		<-senderDone
		t.Log("stopping exchange")
		if err := exchange.Stop(errStopped); err != nil {
			t.Fatalf("error while stopping exchange: %s", err)
		}
	})

	t.Run("Receive|Respond", func(t *testing.T) {
		t.Parallel()
		<-exchangeReady

		receiver = "RECEIVER"
		t.Logf("starting receiver")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		endpoint, err := exchange.OpenCtx(ctx, nil, 1, receiver)
		if err != nil {
			t.Errorf("unexpected error while setting up receiver: %s", err)
			receiverReady <- err
			close(receiverError)
			close(receiverDone)
			return
		}
		defer endpoint.Close(fmt.Errorf("reciever closed"))

		func() {
			t.Logf("receiving")
			defer close(receiverDone)
			if endpoint.Name != receiver {
				err := fmt.Errorf("invalid receiver name: %s, expected %s", endpoint.Name, receiver)
				t.Error(err)
				close(receiverError)
				return
			}
			close(receiverReady)
			header, data, closed, err := endpoint.Receive()
			if err != nil {
				t.Errorf("unexpected error while awaiting message on %s: %s", receiver, err)
				return
			}
			if message != data {
				t.Errorf("unexpected message: %s, expected %s", data, message)
				return
			}
			if header.To != receiver {
				t.Errorf("unexpected to: %s, expected %s", header.To, receiver)
				return
			}
			if closed {
				t.Errorf("unexpected close")
				return
			}
			select {
			case <-header.Lifespan.Done():
				t.Errorf("unexpected end of life for message sent by %s", header.From)
				return
			default:
				t.Logf("responding to %s", header.From)
				if err := endpoint.Send(context.Background(), header.From, fmt.Sprintf("Hello, %s!", header.From)); err != nil {
					t.Errorf("unexpected error while sending message to %s: %s", header.From, err)
				}
			}
		}()

		_, _, closed, err := endpoint.Receive()
		if err != nil {
			if err != errStopped {
				t.Errorf("unexpected error while waiting for stop on %s", receiver)
			}
		}
		if !closed {
			t.Errorf("expected close")
		}
		t.Log("stopped receiver")
	})

	t.Run("Send", func(t *testing.T) {
		t.Parallel()
		<-exchangeReady

		select {
		case <-receiverError:
			return
		case <-receiverReady:
		}

		t.Logf("starting sender")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		endpoint, err := exchange.OpenWildcardCtx(ctx, nil, 1, "*", 16)
		if err != nil {
			t.Errorf("unexpected error while setting up sender: %s", err)
			close(senderDone)
			return
		}
		defer endpoint.Close(fmt.Errorf("sender closed"))

		func() {
			t.Logf("sending")
			defer close(senderDone)
			t.Logf("sending %s from %s to %s", message, endpoint.Name, receiver)
			if err := endpoint.SendCtx(ctx, ctx, receiver, message); err != nil {
				t.Errorf("unexpected error while sending message from %s: %s", endpoint.Name, err)
				return
			}
			t.Logf("awaiting respond")
			_, _, _, err := endpoint.ReceiveCtx(ctx)
			if err != nil {
				t.Errorf("unexpected error while awaiting message on %s: %s", endpoint.Name, err)
				return
			}
		}()

		_, _, closed, err := endpoint.Receive()
		if err != nil {
			if err != errStopped {
				t.Errorf("unexpected error while waiting for stop on %s", endpoint.Name)
			}
		}
		if !closed {
			t.Errorf("expected close")
		}
		t.Log("stopped sender")
	})

}
