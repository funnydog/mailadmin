package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/funnydog/mailadmin/core"
	"github.com/funnydog/mailadmin/core/config"
	"github.com/funnydog/mailadmin/types"
	"github.com/gorilla/csrf"
)

// path to the sqlite3 database
const databasePath = "/tmp/testing.db"

var testingClient = http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func showResponseBody(t *testing.T, res *http.Response) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	t.Error(string(body))
}

func testGet(t *testing.T, url string, status int) {
	res, err := testingClient.Get(url)
	if err != nil {
		panic(err)
	}
	if res.StatusCode != status {
		t.Errorf("Actual status: (%d); Expected status: (%d)",
			res.StatusCode, status)
		showResponseBody(t, res)
	}
}

func testPost(t *testing.T, url, data string, status int) {
	res, err := testingClient.Post(
		url,
		"application/x-www-form-urlencoded",
		strings.NewReader(data),
	)
	if err != nil {
		panic(err)
	}
	if res.StatusCode != status {
		t.Errorf("Actual status: (%d); Excpected status: (%d)",
			res.StatusCode, status)
		showResponseBody(t, res)
	}
}

func createTestingContext() *core.Context {
	conf, err := config.Read("config.json")
	if err != nil {
		panic(err)
	}

	conf.DBType = "sqlite3"
	conf.DBName = databasePath

	ctx, err := core.CreateContextFromConf(&conf)
	if err != nil {
		panic(err)
	}
	configureContext(ctx)

	// skip the CSRF token check
	ctx.AddMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				r = csrf.UnsafeSkipCheck(r)
				h.ServeHTTP(w, r)
			})
	})

	// build the model
	if err = types.CreateModel(ctx.Database); err != nil {
		panic(err)
	}

	// prepare the statements
	if err = types.PrepareStatements(ctx.Database); err != nil {
		panic(err)
	}

	// add some dummy db entries
	domain := types.Domain{
		Name:        "example.com",
		Description: "This domain is an example",
		BackupMX:    false,
		Active:      true,
	}
	err = domain.Create(ctx.Database)
	if err != nil {
		panic(err)
	}

	mailbox := types.Mailbox{
		Domain:   domain.Id,
		Email:    "test@example.com",
		Password: "$6$blah$blah",
		Active:   true,
	}
	err = mailbox.Create(ctx.Database)
	if err != nil {
		panic(err)
	}

	alias := types.Alias{
		Domain:      domain.Id,
		Source:      "postmaster@example.com",
		Destination: "test@example.com",
		Active:      true,
	}
	err = alias.Create(ctx.Database)
	if err != nil {
		panic(err)
	}

	return ctx
}

func closeTestingContext(ctx *core.Context) {
	ctx.Close()
	_ = os.Remove(databasePath)
}

func TestIndexHandler(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	testGet(t, ts.URL+ctx.Reverse("index"), http.StatusMovedPermanently)
}

func TestSignInHandler(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	myURL := ts.URL + ctx.Reverse("sign-in")

	// sign in front page
	testGet(t, myURL, http.StatusOK)

	// failed sign in post
	testPost(t, myURL, "", http.StatusOK)

	// successful sign in post
	data := url.Values{}
	data.Set("username", "admin")
	data.Set("password", "pass")
	testPost(t, myURL, data.Encode(), http.StatusFound)
}

func TestSignOutHandler(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	testGet(t, ts.URL+ctx.Reverse("sign-out"), http.StatusFound)
}

func TestDomainList(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	testGet(t, ts.URL+ctx.Reverse("domain-list"), http.StatusOK)
}

func TestDomainOverview(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	testGet(t, ts.URL+ctx.Reverse("domain-overview", 1), http.StatusOK)
}

func TestDomainCreate(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	myURL := ts.URL + ctx.Reverse("domain-create")

	// create - GET
	testGet(t, myURL, http.StatusOK)

	// create - POST with wrong parameters
	data := url.Values{}
	data.Add("name", "")
	data.Add("description", "a description")
	data.Add("backupmx", "")
	data.Add("active", "on")
	testPost(t, myURL, data.Encode(), http.StatusOK)

	// create - POST with correct parameters
	data.Set("name", "adomain.com")
	testPost(t, myURL, data.Encode(), http.StatusFound)

	// check if the object stored in the db matches the inserted data
	domain, err := types.GetDomainById(ctx.Database, 2)
	if err != nil {
		t.Errorf("Object not found")
	}
	if domain.Name != data.Get("name") {
		t.Errorf("Domain name found: %s, Expected %s",
			domain.Name, data.Get("name"))
	}
	if domain.Description != data.Get("description") {
		t.Errorf("Domain description found: %s, Expected %s",
			domain.Description, data.Get("description"))
	}
	if domain.Active != true {
		t.Error("Domain is not active, Expected active")
	}
	if domain.BackupMX == true {
		t.Error("Domain is BackupMX, Expected not BackupMX")
	}
}

func TestDomainUpdate(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	myURL := ts.URL + ctx.Reverse("domain-update", 1)

	testGet(t, myURL, http.StatusOK)

	data := url.Values{}
	data.Add("name", "")
	data.Add("description", "another description")
	data.Add("backupmx", "on")
	data.Add("active", "")
	testPost(t, myURL, data.Encode(), http.StatusOK)

	data.Set("name", "anothername.com")
	testPost(t, myURL, data.Encode(), http.StatusFound)

	// check if the object in the database has been updated
	domain, err := types.GetDomainById(ctx.Database, 1)
	if err != nil {
		t.Errorf("Object not found")
	}

	if domain.Name != data.Get("name") {
		t.Errorf("Domain name found: %s, Expected %s",
			domain.Name, data.Get("name"))
	}

	if domain.Description != data.Get("description") {
		t.Errorf("Domain description found: %s, Expected %s",
			domain.Description, data.Get("description"))
	}

	if domain.Active != false {
		t.Error("Domain is active, Expected not active")
	}

	if domain.BackupMX != true {
		t.Error("Domain is not BackupMX, Expected BackupMX")
	}
}

func TestDomainDelete(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	myURL := ts.URL + ctx.Reverse("domain-delete", 1)

	testGet(t, myURL, http.StatusOK)

	testPost(t, myURL, "", http.StatusFound)


	// check if the object is still there
	_, err := types.GetDomainById(ctx.Database, 1)
	if err == nil {
		t.Errorf("The Domain hasn't been deleted")
	}
}
