package web

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/naeimc/ciphect/internal/website/logger"
	"github.com/naeimc/ciphect/logging"
)

var ErrInvalidTLSConfiguration = errors.New("invalid TLS configuration")

type Server struct {
	Address     string
	Handler     http.Handler
	Certificate string
	Key         string

	HaltSignals []os.Signal
	StopSignals []os.Signal
	StopTimeout time.Duration

	OnShutdown []func()

	Server *http.Server

	close chan any
}

func (server *Server) ListenAndServe() (err error) {

	if !server.ValidTLS() {
		return ErrInvalidTLSConfiguration
	}

	server.close = make(chan any)

	if server.StopSignals != nil {
		go server.handleStopSignals()
	}

	if server.HaltSignals != nil {
		go server.handleHaltSignal()
	}

	if server.Server == nil {
		server.Server = new(http.Server)
	}
	server.Server.Addr = server.Address
	server.Server.Handler = server.Handler
	for _, shutdown := range server.OnShutdown {
		server.Server.RegisterOnShutdown(shutdown)
	}

	logger.Printf(logging.Notification, "starting server: %s", server.Address)
	if server.UseTLS() {
		err = server.Server.ListenAndServeTLS(server.Certificate, server.Key)
	} else {
		logger.Print(logging.Warning, "TLS/HTTPS is OFF")
		err = server.Server.ListenAndServe()
	}

	if err == http.ErrServerClosed {
		err = nil
		<-server.close
	}

	if err != nil {
		logger.Print(logging.Fatal, err)
	}

	return err
}

func (server *Server) handleStopSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, server.StopSignals...)
	logger.Printf(logging.Notification, "stopping server: os signal received: %s", <-signals)
	if err := server.Stop(); err != nil {
		logger.Printf(logging.Error, "error on stop: %s", err)
	}
	close(server.close)
}

func (server *Server) handleHaltSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, server.HaltSignals...)
	logger.Printf(logging.Notification, "halting server: os signal received: %s", <-signals)
	if err := server.Halt(); err != nil {
		logger.Printf(logging.Error, "error on halt: %s", err)
	}
	close(server.close)
}

func (server *Server) Stop() error {
	ctx := context.Background()
	if server.StopTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, server.StopTimeout)
		defer cancel()
	}
	return server.StopCtx(ctx)
}

func (server *Server) StopCtx(ctx context.Context) error {
	return server.Server.Shutdown(ctx)
}

func (server *Server) Halt() error {
	return server.Server.Close()
}

func (server *Server) UseTLS() bool {
	return (server.Certificate != "" && server.Key != "")
}

func (server *Server) ValidTLS() bool {
	return (server.Certificate == "" && server.Key == "") || (server.Certificate != "" && server.Key != "")
}
