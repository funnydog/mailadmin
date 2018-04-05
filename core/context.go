package core

import (
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"

	"github.com/funnydog/mailadmin/core/config"
	"github.com/funnydog/mailadmin/core/db"
	"github.com/funnydog/mailadmin/core/template"
	"github.com/funnydog/mailadmin/core/urls"
)

type Handler func(http.ResponseWriter, *http.Request, *Context)
type Middleware func(http.Handler) http.Handler

type Context struct {
	Config          *config.Configuration
	Database        *db.Database
	TemplateManager *template.Manager
	URLManager      *urls.Manager
	Router          *httprouter.Router
	Store           *sessions.CookieStore
	Middleware      []Middleware
}

func (c *Context) Close() {
	c.Database.Close()
}

func (c *Context) ExtendAndRender(w http.ResponseWriter, base, template string,
	data *map[string]interface{}) error {

	if data == nil {
		data = &map[string]interface{}{}
	}

	err := c.TemplateManager.Render(w, base, template, data)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (c *Context) Render(w http.ResponseWriter, template string,
	data *map[string]interface{}) error {

	return c.ExtendAndRender(w, "", template, data)
}

func (c *Context) Reverse(name string, args ...interface{}) string {
	url, err := c.URLManager.Reverse(name, args)
	if err != nil {
		log.Fatal(err)
	}
	return url
}

func (c *Context) ListenAndServe() error {
	var router http.Handler = c.Router
	for _, m := range c.Middleware {
		router = m(router)
	}

	server := http.Server{
		Addr:    c.Config.ServerHost + ":" + c.Config.ServerPort,
		Handler: router,
	}

	if c.Config.ServerCert != "" {
		return server.ListenAndServeTLS(
			c.Config.ServerCert,
			c.Config.ServerKey,
		)
	} else {
		return server.ListenAndServe()
	}
}

func embedCtx(fn func(http.ResponseWriter, *http.Request, *Context),
	ctx *Context) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, ctx)
	}
}

func (c *Context) AddRoute(name, method, prefix string, handler Handler) {
	route := urls.URL{
		Prefix:      c.Config.BasePrefix + prefix,
		Method:      method,
		HandlerFunc: embedCtx(handler, c),
		Name:        name,
	}
	c.URLManager.Add(&route)
}

func (c *Context) AddMiddleware(mid Middleware) {
	c.Middleware = append(c.Middleware, mid)
}

func CreateContext(configFile string) (Context, error) {
	conf, err := config.Read(configFile)
	if err != nil {
		return Context{}, err
	}

	db, err := db.Connect(&conf)
	if err != nil {
		return Context{}, err
	}

	// default static directory is /static
	if conf.StaticPrefix == "" {
		conf.StaticPrefix = "/static"
	}

	router := httprouter.New()
	if conf.StaticDir != "" {
		router.ServeFiles(
			conf.StaticPrefix+"/*filepath",
			http.Dir(conf.StaticDir),
		)
	}
	urlManager := urls.CreateManager(router)

	templates, err := template.Create(&conf, &urlManager)
	if err != nil {
		db.Close()
		return Context{}, err
	}

	if conf.CookieKey == "" {
		conf.CookieKey = "something-very-secret"
	}

	return Context{
		Config:          &conf,
		Database:        db,
		TemplateManager: &templates,
		URLManager:      &urlManager,
		Store:           sessions.NewCookieStore([]byte(conf.CookieKey)),
		Router:          router,
	}, nil
}
