package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/naeimc/ciphect/internal/website/web/auth"
)

func Respond(status int, writer http.ResponseWriter, request *http.Request, session *auth.Session, response any) (int, error) {

	for key, value := range session.Headers {
		writer.Header().Add(key, value)
	}

	for _, cookie := range session.Cookies {
		http.SetCookie(writer, cookie)
	}

	if response, ok := response.(Response); ok {
		return response.Respond(status, writer, request)
	}

	if status != http.StatusOK {
		writer.WriteHeader(status)
	}

	switch response := response.(type) {
	case int:
		return writer.Write([]byte(http.StatusText(response)))
	case string:
		return writer.Write([]byte(response))
	case []byte:
		return writer.Write(response)
	}

	return 0, nil
}

type Response interface {
	Respond(int, http.ResponseWriter, *http.Request) (int, error)
}

type Redirect struct {
	URL string
}

func (redirect Redirect) Respond(status int, writer http.ResponseWriter, request *http.Request) (int, error) {
	http.Redirect(writer, request, redirect.URL, status)
	return 0, nil
}

type JSON struct {
	Data any
}

func (JSON JSON) Respond(status int, writer http.ResponseWriter, request *http.Request) (int, error) {
	if status != http.StatusOK {
		writer.WriteHeader(status)
	}

	b, err := json.Marshal(JSON.Data)
	if err != nil {
		return 0, err
	}

	return writer.Write(b)
}

type HTMLTemplate struct {
	Template *template.Template
	Data     any
}

func (template HTMLTemplate) Respond(status int, writer http.ResponseWriter, request *http.Request) (int, error) {
	if status != http.StatusOK {
		writer.WriteHeader(status)
	}

	buffer := new(bytes.Buffer)
	if err := template.Template.Execute(writer, template.Data); err != nil {
		return 0, err
	}

	return writer.Write(buffer.Bytes())
}
