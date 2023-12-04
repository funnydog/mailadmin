package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	"github.com/funnydog/mailadmin/core"
	"github.com/funnydog/mailadmin/core/config"
	"github.com/funnydog/mailadmin/core/form"
	"github.com/funnydog/mailadmin/types"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/pborman/getopt/v2"
)

var (
	//go:embed public
	resFS        embed.FS
	helpFlag     = getopt.Bool('h', "display help")
	createFlag   = getopt.Bool('m', "create a new model")
	passwordFlag = getopt.Bool('p', "change the sign-in password")
	configPath   = getopt.String('f', "config.json", "path to the configuration")
)

func getFlashes(w http.ResponseWriter, r *http.Request, s sessions.Store) []interface{} {
	flashes := []interface{}{}
	session, err := s.Get(r, "session")
	if err != nil {
		return flashes
	}

	flashes = session.Flashes()
	session.Save(r, w)

	return flashes
}

func addFlash(w http.ResponseWriter, r *http.Request, s sessions.Store, text string) error {
	session, err := s.Get(r, "session")
	if err != nil {
		return err
	}

	session.AddFlash(text)
	session.Save(r, w)
	return nil
}

func main() {
	getopt.Parse()
	if *helpFlag {
		getopt.PrintUsage(os.Stdout)
		return
	}

	staticConf := config.Static{
		StaticDir:   "public/static",
		TemplateDir: "public/templates",
		TagsDir:     "public/tags",
		ExtendDir:   "public/extend",
	}

	ctx, err := core.CreateContextFromPath(resFS, staticConf, *configPath)
	if err != nil {
		log.Panic(err)
	}
	defer ctx.Close()

	if *passwordFlag {
		fmt.Print("Type the new password: ")
		bytepwd, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Panic(err)
		}

		hashpwd, err := bcrypt.GenerateFromPassword(bytepwd, bcrypt.DefaultCost)
		if err != nil {
			log.Panic(err)
		}

		ctx.Config.Password = string(hashpwd)
		err = ctx.Config.Write("config.json")
		if err != nil {
			log.Panic(err)
		}

		fmt.Print("\nPassword changed\n")
		return
	}

	if *createFlag {
		fmt.Println("Creating the model")
		types.CreateModel(ctx.Database)
		return
	}

	configureContext(ctx)

	err = types.PrepareStatements(ctx.Database)
	if err != nil {
		log.Panic(err)
	}

	err = ctx.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

type route struct {
	prefix  string
	method  string
	handler core.Handler
	name    string
}

func configureContext(ctx *core.Context) {
	ctx.SetNotFoundTemplate("404.html")
	ctx.SetPanicTemplate("500.html")

	routes := []route{
		{"/", "GET", indexHandler, "index"},

		{"/sign-in/", "GET", signInHandler, "sign-in"},
		{"/sign-in/", "POST", signInHandler, ""},
		{"/sign-out/", "GET", signOutHandler, "sign-out"},

		{"/domain/list/", "GET", domainList, "domain-list"},
		{"/domain/create/", "GET", domainSave, "domain-create"},
		{"/domain/create/", "POST", domainSave, ""},
		{"/domain/overview/:pk", "GET", domainOverview, "domain-overview"},
		{"/domain/update/:pk", "GET", domainSave, "domain-update"},
		{"/domain/update/:pk", "POST", domainSave, ""},
		{"/domain/delete/:pk", "GET", domainDelete, "domain-delete"},
		{"/domain/delete/:pk", "POST", domainDelete, ""},

		{"/mailbox/list/:domain", "GET", mailboxList, "mailbox-list"},
		{"/mailbox/create/:domain", "GET", mailboxSave, "mailbox-create"},
		{"/mailbox/create/:domain", "POST", mailboxSave, ""},
		{"/mailbox/update/:domain/:pk", "GET", mailboxSave, "mailbox-update"},
		{"/mailbox/update/:domain/:pk", "POST", mailboxSave, ""},
		{"/mailbox/delete/:domain/:pk", "GET", mailboxDelete, "mailbox-delete"},
		{"/mailbox/delete/:domain/:pk", "POST", mailboxDelete, ""},

		{"/alias/list/:domain", "GET", aliasList, "alias-list"},
		{"/alias/create/:domain", "GET", aliasSave, "alias-create"},
		{"/alias/create/:domain", "POST", aliasSave, ""},
		{"/alias/update/:domain/:pk", "GET", aliasSave, "alias-update"},
		{"/alias/update/:domain/:pk", "POST", aliasSave, ""},
		{"/alias/delete/:domain/:pk", "GET", aliasDelete, "alias-delete"},
		{"/alias/delete/:domain/:pk", "POST", aliasDelete, ""},
	}

	for _, r := range routes {
		ctx.AddRoute(r.name, r.method, r.prefix, r.handler)
	}

	// the order is important
	// from the last executed to the first

	// check the CSRF
	ctx.AddMiddleware(csrf.Protect(
		[]byte("32-byte-long-auth-key"),
		csrf.Secure(false),
		csrf.FieldName("mailadmin-csrf-token"),
	))

	// check if the user is logged
	ctx.AddMiddleware(func(h http.Handler) http.Handler {
		// always allow the sign-in url
		sign_in := ctx.Reverse("sign-in")
		ctx.AddAllowedURL(sign_in)
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if !ctx.IsURLAllowed(r.URL.Path) {
					session, err := ctx.Store.Get(r, "session")
					if err == nil && session.Values["loggedin"] != true {
						http.Redirect(w, r, sign_in, http.StatusFound)
						return
					}
				}
				h.ServeHTTP(w, r)
			})
	})

	// skip the CSRF check
	// ctx.AddMiddleware(func(h http.Handler) http.Handler {
	// 	return http.HandlerFunc(
	// 		func(w http.ResponseWriter, r *http.Request) {
	// 			if r.URL.Path == "/gather/" || r.URL.Path == "/gather" {
	// 				r = csrf.UnsafeSkipCheck(r)
	// 			}
	// 			h.ServeHTTP(w, r)
	// 		})
	// })
}

func indexHandler(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	http.Redirect(
		w,
		r,
		ctx.Reverse("domain-list"),
		http.StatusMovedPermanently,
	)
}

func signInHandler(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	data := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	}
	if r.Method == "GET" {
		// fallthrough
	} else if r.Method != "POST" {
		// not supported
		return
	} else {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := bcrypt.CompareHashAndPassword([]byte(ctx.Config.Password), []byte(password))
		if username == ctx.Config.Username && err == nil {
			session, err := ctx.Store.Get(r, "session")
			if err == nil {
				session.Values["loggedin"] = true
			}
			session.Save(r, w)
			http.Redirect(w, r, ctx.Reverse("index"), http.StatusFound)
			return
		}

		data["Error"] = "Sign in failed, wrong username/password"
	}
	ctx.Render(w, "sign_in.html", &data)
}

func signOutHandler(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	session, err := ctx.Store.Get(r, "session")
	if err == nil {
		session.Values["loggedin"] = false
	}
	session.Save(r, w)
	http.Redirect(
		w,
		r,
		ctx.Reverse("sign-in"),
		302,
	)
}

func domainList(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	domains, err := types.GetDomainList(ctx.Database)
	if err != nil {
		panic(err)
	}

	ctx.ExtendAndRender(w, "layout", "domain_list.html", &map[string]interface{}{
		"DomainCount": len(domains),
		"domaintab":   true,
		"domains":     domains,
		"flashes":     getFlashes(w, r, ctx.Store),
	})
}

func domainOverview(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := ctx.URLManager.GetParams(r)

	pk, err := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if err != nil {
		panic(err)
	}

	domain, err := types.GetDomainById(ctx.Database, pk)
	if err != nil {
		panic(err)
	}

	data := map[string]interface{}{
		"overviewtab": true,
		"domain":      domain,
		"flashes":     getFlashes(w, r, ctx.Store),
	}

	ctx.ExtendAndRender(w, "layout", "domain_overview.html", &data)
}

func domainForm() form.Form {
	myForm := form.Create()
	myForm.Add("name", &form.TextField{Label: "Name", Required: true, MaxLength: 50})
	myForm.Add("description", &form.TextField{Label: "Description"})
	myForm.Add("backupmx", &form.CheckboxField{Label: "BackupMX"})
	myForm.Add("active", &form.CheckboxField{Label: "Active"})
	return myForm
}

func domainSave(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := ctx.URLManager.GetParams(r)

	form := domainForm()
	data := map[string]interface{}{
		"form":           form,
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	domain := types.Domain{}
	pk, pkerr := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if pkerr != nil {
		domain.Active = true
		data["Title"] = "Create A New Domain"
		data["domaintab"] = true
	} else {
		var err error
		domain, err = types.GetDomainById(ctx.Database, pk)
		if err != nil {
			panic(err)
		}
		data["Title"] = "Change The Domain"
		data["updatetab"] = true
	}
	data["domain"] = domain

	if r.Method == "GET" {
		form.SetString("name", domain.Name)
		form.SetString("description", domain.Description)
		form.SetBool("backupmx", domain.BackupMX)
		form.SetBool("active", domain.Active)
	} else if r.Method != "POST" {
		// not supported
		return
	} else if form.Validate(r) {
		domain.Name = form.GetString("name")
		domain.Description = form.GetString("description")
		domain.BackupMX = form.GetBool("backupmx")
		domain.Active = form.GetBool("active")

		var err error
		var flash string
		if pkerr != nil {
			err = domain.Create(ctx.Database)
			flash = "Domain created successfully"
		} else {
			err = domain.Update(ctx.Database)
			flash = "Domain updated successfully"
		}
		if err != nil {
			panic(err)
		}

		_ = addFlash(w, r, ctx.Store, flash)
		http.Redirect(w, r, ctx.Reverse("domain-overview", domain.Id.Int64), http.StatusFound)
		return
	}

	ctx.ExtendAndRender(w, "layout", "domain_form.html", &data)
}

func domainDelete(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := ctx.URLManager.GetParams(r)

	pk, err := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if err != nil {
		panic(err)
	}

	domain, err := types.GetDomainById(ctx.Database, pk)
	if err != nil {
		panic(err)
	}

	if r.Method == "GET" {
		data := map[string]interface{}{
			"Title":          "Delete the Domain",
			"deletetab":      true,
			"domain":         domain,
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		ctx.ExtendAndRender(w, "layout", "domain_delete.html", &data)
	} else if r.Method != "POST" {
		// method not supported
	} else if err := domain.Delete(ctx.Database); err != nil {
		panic(err)
	} else {
		_ = addFlash(w, r, ctx.Store, "Domain deleted successfully")
		http.Redirect(w, r, ctx.Reverse("domain-list"), http.StatusFound)
	}
}

func mailboxList(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := ctx.URLManager.GetParams(r)

	domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
	if err != nil {
		panic(err)
	}

	domain, err := types.GetDomainById(ctx.Database, domain_id)
	if err != nil {
		panic(err)
	}

	mailboxes, err := types.GetMailboxList(ctx.Database, domain_id)
	if err != nil {
		panic(err)
	}

	ctx.ExtendAndRender(w, "layout", "mailbox_list.html", &map[string]interface{}{
		"Title":        "Managed Mailboxes",
		"MailboxCount": len(mailboxes),
		"mailboxtab":   true,
		"mailboxes":    mailboxes,
		"domain":       domain,
		"flashes":      getFlashes(w, r, ctx.Store),
	})
}

func createMailboxForm(pwdRequired bool) form.Form {
	myForm := form.Create()
	myForm.Add("email", &form.EmailField{Label: "E-Mail", Required: true})
	myForm.Add("password", &form.TextField{Label: "Password", Required: pwdRequired})
	myForm.Add("active", &form.CheckboxField{Label: "Active"})
	return myForm
}

func mailboxSave(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := ctx.URLManager.GetParams(r)

	domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
	if err != nil {
		panic(err)
	}

	domain, err := types.GetDomainById(ctx.Database, domain_id)
	if err != nil {
		panic(err)
	}

	var title string
	mailbox := types.Mailbox{}
	pk, pkerr := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if pkerr != nil {
		mailbox.Domain = domain.Id
		mailbox.Active = true
		title = "Create New Mailbox"
	} else {
		mailbox, err = types.GetMailboxById(ctx.Database, pk)
		if err != nil {
			panic(err)
		}
		title = "Change The Mailbox"
	}

	form := createMailboxForm(pkerr != nil)
	data := map[string]interface{}{
		"form":           form,
		"mailboxtab":     true,
		"domain":         domain,
		"Title":          title,
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	if r.Method == "GET" {
		form.SetString("email", mailbox.Email)
		form.SetBool("active", mailbox.Active)
	} else if r.Method != "POST" {
		// not supported
		return
	} else {
		// form validation
		valid := form.Validate(r)
		if email := r.FormValue("email"); !strings.HasSuffix(email, "@"+domain.Name) {
			valid = false
			form.SetError("email", "The address doesn't end with @"+domain.Name)
		}
		if password := r.FormValue("password"); password == "" && pkerr != nil {
			valid = false
			form.SetError("password", "This field cannot be empty")
		}

		// submit
		if valid {
			mailbox.Email = form.GetString("email")
			mailbox.Active = form.GetBool("active")

			if password := form.GetString("password"); password != "" {
				// hash the password
				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					panic(err)
				}
				mailbox.Password = string(hash)
			}

			var flash string
			if pkerr != nil {
				err = mailbox.Create(ctx.Database)
				flash = "Mailbox created successfully"
			} else {
				err = mailbox.Update(ctx.Database)
				flash = "Mailbox updated successfully"
			}
			if err != nil {
				panic(err)
			}

			_ = addFlash(w, r, ctx.Store, flash)
			http.Redirect(w, r, ctx.Reverse("mailbox-list", domain_id), http.StatusFound)
			return
		}
	}
	ctx.ExtendAndRender(w, "layout", "mailbox_form.html", &data)
}

func mailboxDelete(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := ctx.URLManager.GetParams(r)

	pk, err := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if err != nil {
		panic(err)
	}

	mailbox, err := types.GetMailboxById(ctx.Database, pk)
	if err != nil {
		panic(err)
	}

	if r.Method == "GET" {
		domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
		if err != nil {
			panic(err)
		}

		domain, err := types.GetDomainById(ctx.Database, domain_id)
		if err != nil {
			panic(err)
		}

		data := map[string]interface{}{
			"Title":          "Delete the Mailbox",
			"mailboxtab":     true,
			"domain":         domain,
			"mailbox":        mailbox,
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		ctx.ExtendAndRender(w, "layout", "mailbox_delete.html", &data)
	} else if r.Method != "POST" {
		// method not supported
	} else if err := mailbox.Delete(ctx.Database); err != nil {
		panic(err)
	} else {
		_ = addFlash(w, r, ctx.Store, "Mailbox deleted successfully")
		http.Redirect(w, r, ctx.Reverse("mailbox-list", mailbox.Domain.Int64), http.StatusFound)
	}
}

func aliasList(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := ctx.URLManager.GetParams(r)

	domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
	if err != nil {
		panic(err)
	}

	domain, err := types.GetDomainById(ctx.Database, domain_id)
	if err != nil {
		panic(err)
	}

	aliases, err := types.GetAliasList(ctx.Database, domain_id)
	if err != nil {
		panic(err)
	}

	ctx.ExtendAndRender(w, "layout", "alias_list.html", &map[string]interface{}{
		"Title":      "Managed Aliases",
		"AliasCount": len(aliases),
		"aliastab":   true,
		"aliases":    aliases,
		"domain":     domain,
		"flashes":    getFlashes(w, r, ctx.Store),
	})
}

func createAliasForm() form.Form {
	myForm := form.Create()
	myForm.Add("destination", &form.EmailField{Label: "Destination", Required: true})
	myForm.Add("redirect_to", &form.EmailField{Label: "Redirect to", Required: true})
	myForm.Add("active", &form.CheckboxField{Label: "Active"})
	return myForm
}

func aliasSave(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := ctx.URLManager.GetParams(r)

	domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
	if err != nil {
		panic(err)
	}

	domain, err := types.GetDomainById(ctx.Database, domain_id)
	if err != nil {
		panic(err)
	}

	var title string
	alias := types.Alias{}
	pk, pkerr := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if pkerr != nil {
		alias.Domain = domain.Id
		alias.Active = true
		title = "Create New Alias"
	} else {
		alias, err = types.GetAliasById(ctx.Database, pk)
		if err != nil {
			panic(err)
		}
		title = "Change The Alias"
	}

	form := createAliasForm()
	data := map[string]interface{}{
		"form":           form,
		"aliastab":       true,
		"domain":         domain,
		"Title":          title,
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	if r.Method == "GET" {
		form.SetString("destination", alias.Destination)
		form.SetString("redirect_to", alias.RedirectTo)
		form.SetBool("active", alias.Active)
	} else if r.Method != "POST" {
		// not supported
		return
	} else {
		valid := form.Validate(r)
		if dest := r.FormValue("destination"); !strings.HasSuffix(dest, "@"+domain.Name) {
			valid = false
			form.SetError("destination", "The address doesn't end with @"+domain.Name)
		}

		if valid {
			alias.Destination = form.GetString("destination")
			alias.RedirectTo = form.GetString("redirect_to")
			alias.Active = form.GetBool("active")

			var flash string
			if pkerr != nil {
				err = alias.Create(ctx.Database)
				flash = "Alias created successfully"
			} else {
				err = alias.Update(ctx.Database)
				flash = "Alias updated successfully"
			}
			if err != nil {
				panic(err)
			}

			_ = addFlash(w, r, ctx.Store, flash)
			http.Redirect(w, r, ctx.Reverse("alias-list", domain_id), http.StatusFound)
			return
		}
	}

	ctx.ExtendAndRender(w, "layout", "alias_form.html", &data)
}

func aliasDelete(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := ctx.URLManager.GetParams(r)

	pk, err := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if err != nil {
		panic(err)
	}

	alias, err := types.GetAliasById(ctx.Database, pk)
	if err != nil {
		panic(err)
	}

	if r.Method == "GET" {
		domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
		if err != nil {
			panic(err)
		}

		domain, err := types.GetDomainById(ctx.Database, domain_id)
		if err != nil {
			panic(err)
		}

		data := map[string]interface{}{
			"Title":          "Delete the Alias",
			"aliastab":       true,
			"domain":         domain,
			"alias":          alias,
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		ctx.ExtendAndRender(w, "layout", "alias_delete.html", &data)
	} else if r.Method != "POST" {
		// not supported
	} else if err := alias.Delete(ctx.Database); err != nil {
		panic(err)
	} else {
		_ = addFlash(w, r, ctx.Store, "Alias deleted successfully")
		http.Redirect(w, r, ctx.Reverse("alias-list", alias.Domain.Int64), http.StatusFound)
	}
}
