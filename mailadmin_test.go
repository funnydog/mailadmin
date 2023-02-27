package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/funnydog/mailadmin/core"
	"github.com/funnydog/mailadmin/core/config"
	"github.com/funnydog/mailadmin/types"
	"github.com/gorilla/csrf"
)

// path to the sqlite3 database
const (
	databasePath  = "/tmp/testing.db"
	dummyUsername = "admin"
	dummyPassword = "pass"
)

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
	staticConf := config.Static{
		StaticDir:   "public/static",
		TemplateDir: "public/templates",
		TagsDir:     "public/tags",
		ExtendDir:   "public/extend",
	}

	conf, err := config.Read("config.json")
	if err != nil {
		panic(err)
	}
	conf.DBType = "sqlite3"
	conf.DBName = databasePath

	ctx, err := core.CreateContextFromConf(resFS, staticConf, &conf)
	if err != nil {
		panic(err)
	}

	// update the password with a dummy one
	password, err := bcrypt.GenerateFromPassword([]byte(dummyPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	ctx.Config.Password = string(password)

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
		Password: "$2y$05$wkmccwcQT8JHTSY5mZErKedBBUrJmW39gTk2Toi7qPIF6.dc6prli",
		Active:   true,
	}
	err = mailbox.Create(ctx.Database)
	if err != nil {
		panic(err)
	}

	alias := types.Alias{
		Domain:      domain.Id,
		Destination: "postmaster@example.com",
		RedirectTo:  "test@example.com",
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
	data.Set("username", dummyUsername)
	data.Set("password", dummyPassword)
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
		t.Error("The Domain hasn't been deleted")
	}
}

func TestMailboxCreate(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	myURL := ts.URL + ctx.Reverse("mailbox-create", 1)

	testGet(t, myURL, http.StatusOK)

	data := url.Values{}
	data.Add("email", "notvalidemail")
	data.Add("password", dummyPassword)
	data.Add("active", "on")
	testPost(t, myURL, data.Encode(), http.StatusOK)

	data.Set("email", "valid@example.com")
	testPost(t, myURL, data.Encode(), http.StatusFound)

	// check if the mailbox was inserted
	mailbox, err := types.GetMailboxById(ctx.Database, 2)
	if err != nil {
		t.Error("The mailbox hasn't been created")
	}
	if mailbox.Email != data.Get("email") {
		t.Errorf("The email %s doesnt match the submitted data %s",
			mailbox.Email, data.Get("email"))
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(mailbox.Password),
		[]byte(data.Get("password")),
	)
	if err != nil {
		t.Error(err)
	}

	if mailbox.Active != true {
		t.Error("The mailbox is not active")
	}
}

func TestMailboxUpdate(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	myURL := ts.URL + ctx.Reverse("mailbox-update", 1, 1)

	testGet(t, myURL, http.StatusOK)

	mailbox, err := types.GetMailboxById(ctx.Database, 1)
	if err != nil {
		t.Error("Mailbox not found")
	}
	oldpass := mailbox.Password

	data := url.Values{}
	data.Add("email", "another@example.org")
	data.Add("active", "on")

	testPost(t, myURL, data.Encode(), http.StatusOK)

	data.Set("email", "another@example.com")
	testPost(t, myURL, data.Encode(), http.StatusFound)

	mailbox, err = types.GetMailboxById(ctx.Database, 1)
	if err != nil {
		t.Error("Mailbox not found")
	}

	if mailbox.Email != data.Get("email") {
		t.Errorf("Email %s not updated to %s", mailbox.Email, data.Get("email"))
	}

	if mailbox.Password != oldpass {
		t.Error("Empty password updated the password in the db")
	}

	if mailbox.Active != true {
		t.Error("The mailbox is not active")
	}

	data.Add("password", "12345")
	testPost(t, myURL, data.Encode(), http.StatusFound)

	mailbox, err = types.GetMailboxById(ctx.Database, 1)
	if err != nil {
		t.Error("Mailbox not found")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(mailbox.Password),
		[]byte(data.Get("password")),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestMailboxDelete(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	myURL := ts.URL + ctx.Reverse("mailbox-delete", 1, 1)

	testGet(t, myURL, http.StatusOK)

	testPost(t, myURL, "", http.StatusFound)

	// check if the object is still there
	_, err := types.GetMailboxById(ctx.Database, 1)
	if err == nil {
		t.Error("The Mailbox hasn't been deleted")
	}
}

func TestAliasCreate(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	myURL := ts.URL + ctx.Reverse("alias-create", 1)

	testGet(t, myURL, http.StatusOK)

	data := url.Values{}
	data.Add("destination", "original@example.org")
	data.Add("redirect_to", "redirected@otherdomain.com")
	data.Add("active", "on")
	testPost(t, myURL, data.Encode(), http.StatusOK)

	data.Set("destination", "original@example.com")
	testPost(t, myURL, data.Encode(), http.StatusFound)

	alias, err := types.GetAliasById(ctx.Database, 2)
	if err != nil {
		t.Error(err)
	}

	if alias.Destination != data.Get("destination") {
		t.Errorf("The alias destination %s doesn't match the submitted destination %s",
			alias.Destination, data.Get("destination"))
	}

	if alias.RedirectTo != data.Get("redirect_to") {
		t.Errorf("The alias redirect_to %s doesn't match the submitted redirect_to %s",
			alias.RedirectTo, data.Get("redirect_to"))
	}

	if alias.Active != true {
		t.Error("The alias is not active")
	}
}

func TestAliasUpdate(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	myURL := ts.URL + ctx.Reverse("alias-update", 1, 1)

	testGet(t, myURL, http.StatusOK)

	data := url.Values{}
	data.Add("destination", "original@example.org")
	data.Add("redirect_to", "redirected@otherdomain.com")
	data.Add("active", "on")
	testPost(t, myURL, data.Encode(), http.StatusOK)

	data.Set("destination", "original@example.com")
	testPost(t, myURL, data.Encode(), http.StatusFound)

	alias, err := types.GetAliasById(ctx.Database, 1)
	if err != nil {
		t.Error(err)
	}

	if alias.Destination != data.Get("destination") {
		t.Errorf("The alias destination %s doesn't match the submitted destination %s",
			alias.Destination, data.Get("destination"))
	}

	if alias.RedirectTo != data.Get("redirect_to") {
		t.Errorf("The alias redirect_to %s doesn't match the submitted redirect_to %s",
			alias.RedirectTo, data.Get("redirect_to"))
	}

	if alias.Active != true {
		t.Error("The alias is not active")
	}
}

func TestAliasDelete(t *testing.T) {
	ctx := createTestingContext()
	defer closeTestingContext(ctx)

	ts := httptest.NewServer(ctx.Router)
	defer ts.Close()

	myURL := ts.URL + ctx.Reverse("alias-delete", 1, 1)

	testGet(t, myURL, http.StatusOK)

	testPost(t, myURL, "", http.StatusFound)

	_, err := types.GetAliasById(ctx.Database, 1)
	if err == nil {
		t.Error("The Alias hasn't been deleted")
	}
}
