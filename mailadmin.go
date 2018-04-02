package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/funnydog/mailadmin/core"
	"github.com/funnydog/mailadmin/core/form"
	"github.com/funnydog/mailadmin/types"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"

	"golang.org/x/crypto/bcrypt"
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

type route struct {
	prefix  string
	method  string
	handler core.Handler
	name    string
}

func main() {
	context, err := core.CreateContext("config.json")
	if err != nil {
		log.Panic(err)
	}
	defer context.Close()

	err = types.RegisterDatabase(context.Database)
	if err != nil {
		log.Panic(err)
	}

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
		context.AddRoute(r.name, r.method, r.prefix, r.handler)
	}
	context.Router.ServeFiles("/static/*filepath", http.Dir(context.Config.StaticDir))

	// the order is important
	// from the last executed to the first

	// check the CSRF
	context.AddMiddleware(csrf.Protect(
		[]byte("32-byte-long-auth-key"),
		csrf.Secure(false),
		csrf.FieldName("mailadmin-csrf-token"),
	))

	// check if the user is logged
	context.AddMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/sign-in/" && r.URL.Path != "/static/signin.css" {
					session, err := context.Store.Get(r, "session")
					if err == nil && session.Values["loggedin"] != true {
						http.Redirect(w, r, "/sign-in/", 302)
						return
					}
				}
				h.ServeHTTP(w, r)
			})
	})

	// skip the CSRF check
	// context.AddMiddleware(func(h http.Handler) http.Handler {
	// 	return http.HandlerFunc(
	// 		func(w http.ResponseWriter, r *http.Request) {
	// 			if r.URL.Path == "/gather/" || r.URL.Path == "/gather" {
	// 				r = csrf.UnsafeSkipCheck(r)
	// 			}
	// 			h.ServeHTTP(w, r)
	// 		})
	// })

	context.ListenAndServe()
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
	signin, err := template.ParseFiles("public/templates/sign_in.html")
	if err != nil {
		log.Println(err)
		return
	}

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
			http.Redirect(w, r, ctx.Reverse("index"), http.StatusMovedPermanently)
			return
		}

		data["Error"] = "Sign in failed, wrong username/password"
	}
	signin.Execute(w, data)
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
		log.Println(err)
		return
	}

	ctx.Render(w, "domain_list.html", &map[string]interface{}{
		"domains": domains,
		"flashes": getFlashes(w, r, ctx.Store),
	})
}

func domainOverview(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := r.Context().Value(0).(httprouter.Params)

	pk, err := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	domain, err := types.GetDomainById(ctx.Database, pk)
	if err != nil {
		log.Println(err)
		return
	}

	data := map[string]interface{}{
		"domain":  domain,
		"flashes": getFlashes(w, r, ctx.Store),
	}

	ctx.Render(w, "domain_overview.html", &data)
}

func domainForm() form.Form {
	myForm := form.Create()
	myForm.Add("name", &form.TextField{Label: "Name", Required: true})
	myForm.Add("description", &form.TextField{Label: "Description"})
	myForm.Add("backupmx", &form.CheckboxField{Label: "BackupMX"})
	myForm.Add("active", &form.CheckboxField{Label: "Active"})
	return myForm
}

func domainSave(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := r.Context().Value(0).(httprouter.Params)

	var title string
	domain := types.Domain{}
	pk, pkerr := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if pkerr != nil {
		domain.Active = true
		title = "Create New Domain"
	} else {
		var err error
		domain, err = types.GetDomainById(ctx.Database, pk)
		if err != nil {
			log.Println(err)
			return
		}
		title = "Change The Domain"
	}

	myForm := domainForm()
	data := map[string]interface{}{
		"PK":             pk,
		"Title":          title,
		"form":           myForm,
		"domain":         domain,
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	if r.Method == "GET" {
		myForm.SetString("name", domain.Name)
		myForm.SetString("description", domain.Description)
		myForm.SetBool("backupmx", domain.BackupMX)
		myForm.SetBool("active", domain.Active)
	} else if r.Method != "POST" {
		// not supported
		return
	} else if myForm.Validate(r) {
		domain.Name = myForm.GetString("name")
		domain.Description = myForm.GetString("description")
		domain.BackupMX = myForm.GetBool("backupmx")
		domain.Active = myForm.GetBool("active")

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
			log.Println(err)
			return
		}

		_ = addFlash(w, r, ctx.Store, flash)
		http.Redirect(w, r, ctx.Reverse("domain-overview", domain.Id.Int64), http.StatusSeeOther)
		return
	}

	ctx.Render(w, "domain_form.html", &data)
}

func domainDelete(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := r.Context().Value(0).(httprouter.Params)

	pk, err := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	domain, err := types.GetDomainById(ctx.Database, pk)
	if err != nil {
		log.Println(err)
		return
	}

	if r.Method == "GET" {
		data := map[string]interface{}{
			"domain":         domain,
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		ctx.Render(w, "domain_delete.html", &data)
	} else if r.Method != "POST" {
		// method not supported
	} else if err := domain.Delete(ctx.Database); err != nil {
		log.Println(err)
	} else {
		_ = addFlash(w, r, ctx.Store, "Domain deleted successfully")
		http.Redirect(w, r, ctx.Reverse("domain-list"), http.StatusSeeOther)
	}
}

func mailboxList(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := r.Context().Value(0).(httprouter.Params)

	domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	domain, err := types.GetDomainById(ctx.Database, domain_id)
	if err != nil {
		log.Println(err)
		return
	}

	mailboxes, err := types.GetMailboxList(ctx.Database, domain_id)
	if err != nil {
		log.Println(err)
		return
	}

	ctx.Render(w, "mailbox_list.html", &map[string]interface{}{
		"mailboxes": mailboxes,
		"domain":    domain,
		"flashes":   getFlashes(w, r, ctx.Store),
	})
}

func createMailboxForm() form.Form {
	myForm := form.Create()
	myForm.Add("email", &form.EmailField{Label: "E-Mail", Required: true})
	myForm.Add("password", &form.TextField{Label: "Password", Required: false})
	myForm.Add("active", &form.CheckboxField{Label: "Active"})
	return myForm
}

func mailboxSave(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := r.Context().Value(0).(httprouter.Params)

	domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	domain, err := types.GetDomainById(ctx.Database, domain_id)
	if err != nil {
		log.Println(err)
		return
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
			log.Println(err)
			return
		}
		title = "Change The Mailbox"
	}

	form := createMailboxForm()
	data := map[string]interface{}{
		"PK":             pk,
		"form":           form,
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
	} else if !form.Validate(r) {
		// fallthrough
	} else if password := form.GetString("password"); password == "" && pkerr != nil {
		// the combination of values is not valid
		form.SetError("password", "This field cannot be empty")
	} else if email := form.GetString("email"); !strings.HasSuffix(email, "@"+domain.Name) {
		form.SetError("email", "This email doesn't end with @"+domain.Name)
	} else {
		mailbox.Email = email
		mailbox.Active = form.GetBool("active")

		// hash the password
		if password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
			if err != nil {
				log.Println(err)
				return
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
			log.Println(err)
			return
		}

		_ = addFlash(w, r, ctx.Store, flash)
		http.Redirect(w, r, ctx.Reverse("mailbox-list", domain_id), http.StatusSeeOther)
		return
	}
	ctx.Render(w, "mailbox_form.html", &data)
}

func mailboxDelete(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := r.Context().Value(0).(httprouter.Params)

	pk, err := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	mailbox, err := types.GetMailboxById(ctx.Database, pk)
	if err != nil {
		log.Println(err)
		return
	}

	if r.Method == "GET" {
		domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
		if err != nil {
			log.Println(err)
			return
		}

		domain, err := types.GetDomainById(ctx.Database, domain_id)
		if err != nil {
			log.Println(err)
			return
		}

		data := map[string]interface{}{
			"domain":         domain,
			"mailbox":        mailbox,
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		ctx.Render(w, "mailbox_delete.html", &data)
	} else if r.Method != "POST" {
		// method not supported
	} else if err := mailbox.Delete(ctx.Database); err != nil {
		log.Println(err)
	} else {
		_ = addFlash(w, r, ctx.Store, "Mailbox deleted successfully")
		http.Redirect(w, r, ctx.Reverse("mailbox-list", mailbox.Domain.Int64), http.StatusSeeOther)
	}
}

func aliasList(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := r.Context().Value(0).(httprouter.Params)

	domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	domain, err := types.GetDomainById(ctx.Database, domain_id)
	if err != nil {
		log.Println(err)
		return
	}

	aliases, err := types.GetAliasList(ctx.Database, domain_id)
	if err != nil {
		log.Println(err)
		return
	}

	ctx.Render(w, "alias_list.html", &map[string]interface{}{
		"aliases": aliases,
		"domain":  domain,
		"flashes": getFlashes(w, r, ctx.Store),
	})
}

func createAliasForm() form.Form {
	myForm := form.Create()
	myForm.Add("source", &form.EmailField{Label: "Source", Required: true})
	myForm.Add("destination", &form.EmailField{Label: "Destination", Required: true})
	myForm.Add("active", &form.CheckboxField{Label: "Active"})
	return myForm
}

func aliasSave(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := r.Context().Value(0).(httprouter.Params)

	domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	domain, err := types.GetDomainById(ctx.Database, domain_id)
	if err != nil {
		log.Println(err)
		return
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
			log.Println(err)
			return
		}
		title = "Change The Alias"
	}

	form := createAliasForm()
	data := map[string]interface{}{
		"PK":             pk,
		"form":           form,
		"domain":         domain,
		"Title":          title,
		csrf.TemplateTag: csrf.TemplateField(r),
	}

	if r.Method == "GET" {
		form.SetString("source", alias.Source)
		form.SetString("destination", alias.Destination)
		form.SetBool("active", alias.Active)
	} else if r.Method != "POST" {
		// not supported
		return
	} else if !form.Validate(r) {
		// fallthrough
	} else if source := form.GetString("source"); !strings.HasSuffix(source, "@"+domain.Name) {
		form.SetError("source", "This source address doesn't end with @"+domain.Name)
	} else {
		alias.Source = source
		alias.Destination = form.GetString("destination")
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
			log.Println(err)
			return
		}

		_ = addFlash(w, r, ctx.Store, flash)
		http.Redirect(w, r, ctx.Reverse("alias-list", domain_id), http.StatusSeeOther)
		return
	}

	ctx.Render(w, "alias_form.html", &data)
}

func aliasDelete(w http.ResponseWriter, r *http.Request, ctx *core.Context) {
	parameters := r.Context().Value(0).(httprouter.Params)

	pk, err := strconv.ParseInt(parameters.ByName("pk"), 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	alias, err := types.GetAliasById(ctx.Database, pk)
	if err != nil {
		log.Println(err)
		return
	}

	if r.Method == "GET" {
		domain_id, err := strconv.ParseInt(parameters.ByName("domain"), 10, 64)
		if err != nil {
			log.Println(err)
			return
		}

		domain, err := types.GetDomainById(ctx.Database, domain_id)
		if err != nil {
			log.Println(err)
			return
		}

		data := map[string]interface{}{
			"domain":         domain,
			"alias":          alias,
			csrf.TemplateTag: csrf.TemplateField(r),
		}

		ctx.Render(w, "alias_delete.html", &data)
	} else if r.Method != "POST" {
		// not supported
	} else if err := alias.Delete(ctx.Database); err != nil {
		log.Println(err)
	} else {
		_ = addFlash(w, r, ctx.Store, "Alias deleted successfully")
		http.Redirect(w, r, ctx.Reverse("alias-list", alias.Id.Int64), http.StatusSeeOther)
	}
}
