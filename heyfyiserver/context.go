package heyfyiserver

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gocraft/web"
	"github.com/gorilla/sessions"
	"github.com/kiwih/heyfyi/heyfyiserver/account"
	"github.com/kiwih/heyfyi/heyfyiserver/fact"
	"github.com/kiwih/heyfyi/heyfyiserver/fyidb"
)

type AnyStorer interface {
	account.AccountStorer
	fact.FactStorer
}

//Used in all requests
type Context struct {
	ErrorMessages        []string
	NotificationMessages []string
	Data                 interface{}
	Store                *sessions.CookieStore
	Account              *account.Account
	Storage              AnyStorer
}

//HELPER FUNCTIONS

//This function allows for a handler to set an error message as a "Flash" message which can be shown to the user in a later request
//(via a different handler) - it stores them in a session variable
func (c *Context) SetErrorMessage(rw web.ResponseWriter, req *web.Request, err string) {
	session, _ := c.Store.Get(req.Request, "error-messages")
	session.AddFlash(err)
	session.Save(req.Request, rw)
}

//This function allows for a handler to set a notification message as a "Flash" message which can be shown to the user in a later request
//(via a different handler) - it stores them in a session variable
func (c *Context) SetNotificationMessage(rw web.ResponseWriter, req *web.Request, notification string) {
	session, _ := c.Store.Get(req.Request, "notification-messages")
	session.AddFlash(notification)
	session.Save(req.Request, rw)
}

//This function allows us to store a bad request from a form (eg not meeting the regex for the NHI parameter of system.Patient)
//so it can be recalled later for them to amend it
func (c *Context) SetFailedRequestObject(rw web.ResponseWriter, req *web.Request, requestedObject interface{}) {
	session, _ := c.Store.Get(req.Request, "error-form-requests")
	session.AddFlash(requestedObject)
	session.Save(req.Request, rw)
}

//This function returns just one "flash" failed request object for a session. Any other request objects that were stored will be
//removed without retrieval. It allows for users to amend bad forms without needing to retype all the data
func (c *Context) CheckFailedRequestObject(rw web.ResponseWriter, req *web.Request) interface{} {
	session, _ := c.Store.Get(req.Request, "error-form-requests")
	flashes := session.Flashes()
	session.Save(req.Request, rw)

	if len(flashes) > 0 {
		//again, note that only the first one is returned. All other forms will be discarded.
		return flashes[0]
	}
	return nil
}

//MIDDLEWARE

func (c *Context) AssignStorageMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	c.Storage = &fyidb.DbStorage
	next(rw, req)
}

func (c *Context) LoadUserMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	session, _ := c.Store.Get(req.Request, "session-security")

	if session.Values["sessionId"] != nil {
		c.Account, _ = c.Storage.LoadAccountFromSession(session.Values["sessionId"].(string))
	}
	next(rw, req)
}

func (c *Context) AssignTemplatesAndSessionsMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	c.Store = store
	next(rw, req)
}

//This function returns any flash error messages that have been saved. Upon retrieving them, they will be deleted from the session
//(as they are "flash" session variables)
func (c *Context) GetErrorMessagesMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	session, _ := c.Store.Get(req.Request, "error-messages")
	flashes := session.Flashes()
	session.Save(req.Request, rw)

	if len(flashes) > 0 {
		//it is not possible in go to cast from []interface to []string
		strings := make([]string, len(flashes))
		for i := range flashes {
			strings[i] = flashes[i].(string)
		}
		c.ErrorMessages = strings
	}
	next(rw, req)
}

//This function returns any flash notification messages that have been saved. Upon retrieving them, they will be deleted from the session
//(as they are "flash" session variables)
func (c *Context) GetNotificationMessagesMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	session, _ := c.Store.Get(req.Request, "notification-messages")
	flashes := session.Flashes()
	session.Save(req.Request, rw)

	if len(flashes) > 0 {
		//it is not possible in go to cast from []interface to []string
		strings := make([]string, len(flashes))
		for i := range flashes {
			strings[i] = flashes[i].(string)
		}
		c.NotificationMessages = strings
	}
	next(rw, req)
}

func (c *Context) RequireAccountMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	if c.Account == nil {
		c.SetErrorMessage(rw, req, "You need to sign in to view this page!")
		http.Redirect(rw, req.Request, "/", http.StatusSeeOther)
	} else {
		next(rw, req)
	}
}

//HANDLERS

func (c *Context) HomeHandler(rw web.ResponseWriter, req *web.Request) {
	err := templates.ExecuteTemplate(rw, "homePage", c)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Context) ViewFactHandler(rw web.ResponseWriter, req *web.Request) {
	factIdStr, ok := req.PathParams["factId"]
	if !ok {
		http.Error(rw, "400: Bad fact ID", http.StatusBadRequest)
		return
	}

	//get the fact ID from the URL
	factId, err := strconv.ParseInt(factIdStr, 10, 64)
	if err != nil {
		http.Error(rw, "400: Bad fact ID", http.StatusBadRequest)
		return
	}

	f, err := c.Storage.LoadFactFromId(factId)
	if err != nil {
		http.Error(rw, "404: Fact not found", http.StatusNotFound)
		return
	}

	if f.AwaitModeration {
		if c.Account == nil {
			http.Error(rw, "404: Fact not found", http.StatusNotFound)
			return
		}
		if f.AccountId != c.Account.Id && !c.Account.Admin {
			http.Error(rw, "404: Fact not found", http.StatusNotFound)
			return
		}
	}

	data := struct {
		Fact *fact.Fact
	}{
		Fact: f,
	}
	c.Data = data
	if err := templates.ExecuteTemplate(rw, "factPage", c); err != nil {
		log.Println("Error:", err.Error())
	}
}

type LoginRequestForm struct {
	Email    string
	Password string
	Remember bool
}

func (c *Context) DoSignInRequestHandler(rw web.ResponseWriter, req *web.Request) {
	req.ParseForm()

	var prop LoginRequestForm
	if err := decoder.Decode(&prop, req.PostForm); err != nil {
		c.SetErrorMessage(rw, req, "Decoding error: "+err.Error())
		http.Redirect(rw, req.Request, SignUpUrl.Make(), http.StatusSeeOther)
		return
	}

	propUser, err := account.AttemptLogin(c.Storage, prop.Email, prop.Password, prop.Remember)

	if propUser != nil {
		//they have passed the login check. Save them to the session and redirect to management portal
		session, _ := c.Store.Get(req.Request, "session-security")
		session.Values["sessionId"] = propUser.CurrentSession.String
		c.SetNotificationMessage(rw, req, "Hi, "+propUser.Nickname+".")
		session.Save(req.Request, rw)
		http.Redirect(rw, req.Request, HomeUrl.Make(), http.StatusFound)
		return
	}
	c.SetErrorMessage(rw, req, err.Error())
	http.Redirect(rw, req.Request, HomeUrl.Make(), http.StatusSeeOther)
}

func (c *Context) ListFactsHandler(rw web.ResponseWriter, req *web.Request) {
	//TODO: not signed in can view all posts?
	var accountId int64 = 0
	listFacts := false
	//if logged in, set to view facts that are theirs
	if c.Account != nil {
		accountId = c.Account.Id
		//only show all facts if they are an admin
		listFacts = c.Account.Admin
	}
	facts, err := c.Storage.ListFacts(accountId, listFacts)
	if err != nil {
		http.Error(rw, "500: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Facts []fact.Fact
	}{
		Facts: facts,
	}

	c.Data = data

	if err := templates.ExecuteTemplate(rw, "listFactsPage", c); err != nil {
		log.Println("Error:", err.Error())
	}
}

type CreateAccount struct {
	Email           string
	Password        string
	ConfirmPassword string
	TermsCB         bool
	Nickname        string
}

//This handler is for the registration of an account
func (c *Context) SignUpHandler(rw web.ResponseWriter, req *web.Request) {
	//make sure we aren't signed in
	if c.Account != nil {
		c.SetErrorMessage(rw, req, "You're already signed in!")
		http.Redirect(rw, req.Request, HomeUrl.Make(), http.StatusSeeOther)
		return
	}

	var u CreateAccount

	badU := c.CheckFailedRequestObject(rw, req)
	if badU != nil {
		u = badU.(CreateAccount)
	}

	c.Data = u

	err := templates.ExecuteTemplate(rw, "signUpPage", c)
	if err != nil {
		http.Error(rw, "500: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Context) DoSignUpHandler(rw web.ResponseWriter, req *web.Request) {

	req.ParseForm()

	var u CreateAccount

	if err := decoder.Decode(&u, req.PostForm); err != nil {
		c.SetErrorMessage(rw, req, "Decoding error: "+err.Error())
		http.Redirect(rw, req.Request, SignUpUrl.Make(), http.StatusSeeOther)
		return
	}

	if u.Password != u.ConfirmPassword {
		c.SetFailedRequestObject(rw, req, u)
		c.SetErrorMessage(rw, req, "Your passwords don't match!")
		http.Redirect(rw, req.Request, SignUpUrl.Make(), http.StatusSeeOther)
		return
	}

	if u.TermsCB != true {
		c.SetFailedRequestObject(rw, req, u)
		c.SetErrorMessage(rw, req, "You must accept the terms and conditions!")
		http.Redirect(rw, req.Request, SignUpUrl.Make(), http.StatusSeeOther)
		return
	}

	if err := account.CheckAndCreateAccount(c.Storage, u.Email, u.Password, u.Nickname); err != nil {
		c.SetFailedRequestObject(rw, req, u)
		c.SetErrorMessage(rw, req, err.Error())
		http.Redirect(rw, req.Request, SignUpUrl.Make(), http.StatusSeeOther)
		return
	}

	c.SetNotificationMessage(rw, req, "Your account has been created. Please wait for your verification email, verify, and then you can sign in!")
	http.Redirect(rw, req.Request, HomeUrl.Make(), http.StatusFound)
}

func (c *Context) DoVerificationRequestHandler(rw web.ResponseWriter, req *web.Request) {
	accountIdStr, ok := req.PathParams["accountId"]
	if !ok {
		http.Error(rw, "400: Bad account ID", http.StatusBadRequest)
		//next(rw, req)
		return
	}

	//get the account ID from the URL
	accountId, err := strconv.ParseInt(accountIdStr, 10, 64)
	if err != nil {
		http.Error(rw, "400: Bad account ID", http.StatusBadRequest)
		return
	}

	a, err := c.Storage.LoadAccountFromId(accountId)
	if err != nil {
		http.Error(rw, "404: Account not found", http.StatusNotFound)
		return
	}

	verificationCode, ok := req.PathParams["verificationCode"]
	if !ok {
		http.Error(rw, "400: Bad verification code", http.StatusBadRequest)
		return
	}

	if err := a.ApplyVerificationCode(c.Storage, verificationCode); err != nil {
		http.Error(rw, "400: "+err.Error(), http.StatusBadRequest)
		return
	}

	c.SetNotificationMessage(rw, req, "Your account has been verified and you may now sign in!")
	http.Redirect(rw, req.Request, HomeUrl.Make(), http.StatusFound)
}

func (c *Context) BeginPasswordResetRequestHandler(rw web.ResponseWriter, req *web.Request) {
	err := templates.ExecuteTemplate(rw, "beginResetPasswordPage", c)
	if err != nil {
		http.Error(rw, "500: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

type PasswordResetRequestForm struct {
	Email string
}

func (c *Context) DoBeginPasswordResetRequestHandler(rw web.ResponseWriter, req *web.Request) {
	req.ParseForm()

	var p PasswordResetRequestForm

	if err := decoder.Decode(&p, req.PostForm); err != nil {
		c.SetErrorMessage(rw, req, "Decoding error: "+err.Error())
		http.Redirect(rw, req.Request, ResetPasswordUrl.Make(), http.StatusSeeOther)
		return
	}

	account.DoPasswordResetRequestIfPossible(c.Storage, p.Email)

	c.SetNotificationMessage(rw, req, "Password reset requested.")
	http.Redirect(rw, req.Request, HomeUrl.Make(), http.StatusFound)

}

func (c *Context) PasswordResetRequestHandler(rw web.ResponseWriter, req *web.Request) {
	accountIdStr, ok := req.PathParams["accountId"]
	if !ok {
		http.Error(rw, "400: Bad account ID", http.StatusBadRequest)
		return
	}

	//get the account ID from the URL
	accountId, err := strconv.ParseInt(accountIdStr, 10, 64)
	if err != nil {
		http.Error(rw, "400: Bad account ID", http.StatusBadRequest)
		return
	}

	a, err := c.Storage.LoadAccountFromId(accountId)
	if err != nil {
		http.Error(rw, "404: Account not found", http.StatusNotFound)
		return
	}

	if !a.AwaitingPasswordReset() {
		http.Error(rw, "400: Bad account ID", http.StatusBadRequest)
		return
	}

	_, ok = req.PathParams["resetVerificationCode"]
	if !ok {
		http.Error(rw, "400: Bad reset verification code", http.StatusBadRequest)
		return
	}

	err = templates.ExecuteTemplate(rw, "resetPasswordPage", c)
	if err != nil {
		http.Error(rw, "500: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

type ResetPasswordForm struct {
	Password        string
	ConfirmPassword string
}

func (c *Context) DoPasswordResetRequestHandler(rw web.ResponseWriter, req *web.Request) {

	req.ParseForm()

	var p ResetPasswordForm

	if err := decoder.Decode(&p, req.PostForm); err != nil {
		c.SetErrorMessage(rw, req, "Decoding error: "+err.Error())
		http.Redirect(rw, req.Request, ResetPasswordUrl.Make(), http.StatusSeeOther)
		return
	}

	accountIdStr, ok := req.PathParams["accountId"]
	if !ok {
		http.Error(rw, "400: Bad account ID", http.StatusBadRequest)
		//next(rw, req)
		return
	}

	resetVerificationCode, ok := req.PathParams["resetVerificationCode"]
	if !ok {
		http.Error(rw, "400: Bad verification code", http.StatusBadRequest)
		return
	}

	if p.Password != p.ConfirmPassword {
		c.SetErrorMessage(rw, req, "Your passwords don't match!")
		http.Redirect(rw, req.Request, ResetPasswordUrl.Make("accountId", accountIdStr, "resetVerificationCode", resetVerificationCode), http.StatusSeeOther)
		return
	}

	//get the account ID from the URL
	accountId, err := strconv.ParseInt(accountIdStr, 10, 64)
	if err != nil {
		http.Error(rw, "400: Bad account ID", http.StatusBadRequest)
		return
	}

	a, err := c.Storage.LoadAccountFromId(accountId)
	if err != nil {
		http.Error(rw, "404: Account not found", http.StatusNotFound)
		return
	}

	if err := a.ApplyPasswordResetVerificationCode(c.Storage, resetVerificationCode, p.Password); err != nil {
		c.SetErrorMessage(rw, req, err.Error())
		http.Redirect(rw, req.Request, ResetPasswordUrl.Make("accountId", accountIdStr, "resetVerificationCode", resetVerificationCode), http.StatusSeeOther)
		return
	}

	c.SetNotificationMessage(rw, req, "Password reset - you may now sign in!")
	http.Redirect(rw, req.Request, HomeUrl.Make(), http.StatusFound)
}
