package urls

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
)

var (
	argNotFound        = errors.New("arg not found in prefix")
	methodNotSupported = errors.New("method not supported")
	splitArg           = regexp.MustCompile(":[^/]+")
)

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

func (m *Manager) Add(url *URL) error {
	switch url.Method {
	case "GET":
		m.router.GET(url.Prefix, embedParams(url.HandlerFunc))

	case "POST":
		m.router.POST(url.Prefix, embedParams(url.HandlerFunc))

	default:
		return methodNotSupported
	}
	if url.Name != "" {
		m.urls[url.Name] = url
		m.reverse[url.Name] = splitArg.Split(url.Prefix, -1)
	}
	return nil
}

func (m *Manager) Reverse(name string, args []interface{}) (string, error) {
	tokens, ok := m.reverse[name]
	if !ok {
		return "", fmt.Errorf("reverse for (%s, %v) not found", name, args)
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
