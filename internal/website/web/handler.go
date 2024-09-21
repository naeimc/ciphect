package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/naeimc/ciphect/internal/website/logger"
	"github.com/naeimc/ciphect/internal/website/web/auth"
	"github.com/naeimc/ciphect/logging"
)

type HandlerFunction func(http.ResponseWriter, *http.Request, *auth.Session) (int, any)

func HandleFunction(multiplexer *http.ServeMux, pattern string, handle HandlerFunction) {
	multiplexer.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		session, err := auth.Authenticate(context.Background(), request)
		if err != nil {
			logger.Print(logging.Error, err)
		}

		status, response := handle(writer, request, session)
		size, err := Respond(status, writer, request, session, response)
		logger.Print(logging.Information, NewLogAccess(request, session.Username(), status, size, err))
	})
}

type LogAccess struct {
	RemoteAddress string `json:"remote_address"`
	User          string `json:"user"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	Protocol      string `json:"protocol"`
	Status        int    `json:"status"`
	Size          int    `json:"size"`
	Error         string `json:"error"`
}

func NewLogAccess(request *http.Request, username string, status, size int, err error) LogAccess {
	var e string
	if err != nil {
		e = err.Error()
	}

	return LogAccess{
		RemoteAddress: request.RemoteAddr,
		User:          username,
		Method:        request.Method,
		Path:          request.URL.Path,
		Protocol:      request.Proto,
		Status:        status,
		Size:          size,
		Error:         e,
	}
}

func (log LogAccess) String() string {
	return fmt.Sprintf("%s %s \"%s %s %s\" %d %d %s", log.RemoteAddress, log.User, log.Method, log.Path, log.Protocol, log.Status, log.Size, log.Error)
}
