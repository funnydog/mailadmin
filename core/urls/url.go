package urls

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
)

var splitArg = regexp.MustCompile(":[^/]+")

type ErrMethodNotSupported string

func (mn ErrMethodNotSupported) Error() string {
	return fmt.Sprintf("Method '%s' not supported", mn)
}

type ErrReverseURLNotFound string

func (rn ErrReverseURLNotFound) Error() string {
	return fmt.Sprintf("Reverse URL for '%s' not found", rn)
}

type ErrPrefixAlreadyInserted string

func (pn ErrPrefixAlreadyInserted) Error() string {
	return fmt.Sprintf("The prefix '%s' is already inserted", pn)
}

type ErrNameAlreadyInserted string

func (nn ErrNameAlreadyInserted) Error() string {
	return fmt.Sprintf("The name '%s' is already inserted", nn)
}

type URL struct {
	Prefix      string
	Method      string
	HandlerFunc http.HandlerFunc
	Name        string
}

type Manager struct {
	router  *httprouter.Router
	urls    map[string]*URL
	reverse map[string][]string
}

func embedParams(next func(http.ResponseWriter, *http.Request)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, 0, p)
		next(w, r.WithContext(ctx))
	}
}

func (m *Manager) Add(url *URL) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrPrefixAlreadyInserted(url.Prefix)
		}
	}()

	_, ok := m.urls[url.Name]
	if ok {
		return ErrNameAlreadyInserted(url.Name)
	}
	switch url.Method {
	case "GET":
		m.router.GET(url.Prefix, embedParams(url.HandlerFunc))

	case "POST":
		m.router.POST(url.Prefix, embedParams(url.HandlerFunc))

	default:
		return ErrMethodNotSupported(url.Method)
	}

	if url.Name != "" {
		m.urls[url.Name] = url
		m.reverse[url.Name] = splitArg.Split(url.Prefix, -1)
	}

	return nil
}

func (m *Manager) GetParams(r *http.Request) httprouter.Params {
	return r.Context().Value(0).(httprouter.Params)
}

func (m *Manager) Reverse(name string, args []interface{}) (string, error) {
	tokens, ok := m.reverse[name]
	if !ok {
		return "", ErrReverseURLNotFound(name)
	}

	var t []string
	for index, value := range tokens {
		t = append(t, value)
		if index < len(args) {
			t = append(t, fmt.Sprintf("%v", args[index]))
		}
	}

	return strings.Join(t, ""), nil
}

func CreateManager(router *httprouter.Router) Manager {
	return Manager{
		router,
		map[string]*URL{},
		map[string][]string{},
	}
}
