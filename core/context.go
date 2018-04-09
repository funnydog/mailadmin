package core

import (
	"log"
	"net/http"

	"github.com/funnydog/mailadmin/core/config"
	"github.com/funnydog/mailadmin/core/db"
	"github.com/funnydog/mailadmin/core/template"
	"github.com/funnydog/mailadmin/core/urls"
	"github.com/go-errors/errors"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
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

func (c *Context) SetNotFoundTemplate(template string) {
	if c.Config.Debug && template != "" {
		c.Router.NotFound = http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				log.Printf("Page not found: %s\n", r.URL.Path)
				data := map[string]interface{}{
					"request": r,
				}
				c.Render(w, template, &data)
			},
		)
	} else {
		c.Router.NotFound = nil
	}
}

func badRequest(w http.ResponseWriter, r *http.Request, err interface{}) {
	http.Error(w, "500 bad request", http.StatusBadRequest)
}

func (c *Context) SetPanicTemplate(template string) {
	if c.Config.Debug && template != "" {
		c.Router.PanicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
			log.Println(err)
			data := map[string]interface{}{
				"error":   errors.Wrap(err, 3),
				"request": r,
			}
			c.Render(w, template, &data)
		}
	} else {
		c.Router.PanicHandler = badRequest
	}
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

func (c *Context) AddMiddleware(mid Middleware) {
	c.Middleware = append(c.Middleware, mid)
}

func CreateContextFromConf(conf *config.Configuration) (*Context, error) {
	db, err := db.Connect(conf)
	if err != nil {
		return nil, err
	}

	// default static directory is /static
	if conf.StaticPrefix == "" {
		conf.StaticPrefix = "/static"
	}

	router := httprouter.New()
	router.PanicHandler = badRequest
	if conf.StaticDir != "" {
		router.ServeFiles(
			conf.StaticPrefix+"/*filepath",
			http.Dir(conf.StaticDir),
		)
	}
	urlManager := urls.CreateManager(router)

	templates, err := template.Create(conf, &urlManager)
	if err != nil {
		db.Close()
		return nil, err
	}

	if conf.CookieKey == "" {
		conf.CookieKey = "something-very-secret"
	}

	return &Context{
		Config:          conf,
		Database:        db,
		TemplateManager: &templates,
		URLManager:      &urlManager,
		Store:           sessions.NewCookieStore([]byte(conf.CookieKey)),
		Router:          router,
	}, nil
}

func CreateContextFromPath(configFile string) (*Context, error) {
	conf, err := config.Read(configFile)
	if err != nil {
		return nil, err
	}

	return CreateContextFromConf(&conf)
}
