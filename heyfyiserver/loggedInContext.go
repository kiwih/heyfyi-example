package heyfyiserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gocraft/web"
	"github.com/kiwih/heyfyi/heyfyiserver/account"
	"github.com/kiwih/heyfyi/heyfyiserver/fact"
)

type LoggedInContext struct {
	*Context
}

//This handler performs the logout request
func (c *LoggedInContext) DoSignOutRequestHandler(rw web.ResponseWriter, req *web.Request) {
	session, _ := c.Store.Get(req.Request, "session-security")
	session.Values["sessionId"] = nil
	c.SetNotificationMessage(rw, req, "Goodbye!")

	c.Account.ExpireSession(c.Storage)

	session.Save(req.Request, rw)
	http.Redirect(rw, req.Request, HomeUrl.Make(), http.StatusFound)
}

func (c *LoggedInContext) VoteOnFactHandler(rw web.ResponseWriter, req *web.Request) {
	voteRequest := struct {
		FactId int64
		Up     bool
	}{}

	response := struct {
		Response    string
		FactId      int64
		NewScore    fact.VoteScore
		NewVoteBank int64
	}{}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&voteRequest); err != nil {
		http.Error(rw, "400: "+err.Error(), http.StatusBadRequest)
		return
	}

	if voteRequest.FactId == 0 {
		http.Error(rw, "400: No FactID specified", http.StatusBadRequest)
		return
	}

	f, err := c.Storage.LoadFactFromId(voteRequest.FactId)
	if err != nil {
		http.Error(rw, "400: Bad FactID specified", http.StatusBadRequest)
		return
	}

	if f.AwaitModeration && !c.Account.Admin {
		if f.AccountId != c.Account.Id {
			http.Error(rw, "400: Bad FactID specified", http.StatusBadRequest)
			return
		}
	}

	if err := c.Account.UpdateVoteBank(c.Storage, voteRequest.Up, f.GetScore(c.Account.Id).AccountVote); err != nil {
		if err == account.NoVotesLeft {
			http.Error(rw, "You have no votes to cast!", http.StatusBadRequest)
			return
		} else {
			http.Error(rw, "500: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if _, err := fact.VoteForFact(c.Storage, c.Account.Id, f.Id, voteRequest.Up); err != nil {
		http.Error(rw, "500: "+err.Error(), http.StatusInternalServerError)
		return
	}

	f, _ = c.Storage.LoadFactFromId(f.Id)

	response.FactId = f.Id
	response.NewScore = f.GetScore(c.Account.Id)
	response.NewVoteBank = c.Account.VoteBank
	response.Response = "ok"
	ReturnJSON(rw, response)
}

func (c *LoggedInContext) ModerateFactHandler(rw web.ResponseWriter, req *web.Request) {
	if !c.Account.Admin {
		http.Error(rw, "400: Only admins can make this request", http.StatusBadRequest)
		return
	}

	moderateRequest := struct {
		FactId int64
		Enable bool
	}{}

	response := struct {
		Response           string
		FactId             int64
		NewAwaitModeration bool
	}{}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&moderateRequest); err != nil {
		http.Error(rw, "400: "+err.Error(), http.StatusBadRequest)
		return
	}

	if moderateRequest.FactId == 0 {
		http.Error(rw, "400: No FactID specified", http.StatusBadRequest)
		return
	}

	f, err := c.Storage.LoadFactFromId(moderateRequest.FactId)
	if err != nil {
		http.Error(rw, "400: Bad FactID specified", http.StatusBadRequest)
		return
	}

	if err := c.Storage.ModerateFact(f, moderateRequest.Enable); err != nil {
		http.Error(rw, "500: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response.FactId = f.Id
	response.NewAwaitModeration = f.AwaitModeration
	response.Response = "ok"
	ReturnJSON(rw, response)
}

func (c *Context) CreateFactHandler(rw web.ResponseWriter, req *web.Request) {

	var f fact.Fact

	badF := c.CheckFailedRequestObject(rw, req)
	if badF != nil {
		f = badF.(fact.Fact)
	}

	c.Data = f

	err := templates.ExecuteTemplate(rw, "createFactPage", c)
	if err != nil {
		http.Error(rw, "500: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Context) DoCreateFactHandler(rw web.ResponseWriter, req *web.Request) {

	req.ParseForm()

	var f fact.Fact

	if err := decoder.Decode(&f, req.PostForm); err != nil {
		c.SetErrorMessage(rw, req, "Decoding error: "+err.Error())
		http.Redirect(rw, req.Request, CreateFactUrl.Make(), http.StatusSeeOther)
		return
	}

	f.AccountId = c.Account.Id

	if err := fact.CreateFact(c.Storage, &f); err != nil {
		c.SetFailedRequestObject(rw, req, f)
		c.SetErrorMessage(rw, req, err.Error())
		http.Redirect(rw, req.Request, CreateFactUrl.Make(), http.StatusSeeOther)
		return
	}

	c.SetNotificationMessage(rw, req, "Fact submitted successfully!")
	http.Redirect(rw, req.Request, ViewFactUrl.Make("factId", strconv.FormatInt(f.Id, 10)), http.StatusFound)
}

func (c *Context) DeleteFactHandler(rw web.ResponseWriter, req *web.Request) {

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

	if f.AccountId != c.Account.Id && !c.Account.Admin {
		http.Error(rw, "400: Bad fact ID", http.StatusBadRequest)
		return
	}

	c.Data = f

	err = templates.ExecuteTemplate(rw, "deleteFactPage", c)
	if err != nil {
		http.Error(rw, "500: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *Context) DoDeleteFactHandler(rw web.ResponseWriter, req *web.Request) {

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

	if f.AccountId != c.Account.Id && !c.Account.Admin {
		http.Error(rw, "400: Bad fact ID", http.StatusBadRequest)
		return
	}

	if err := c.Storage.DeleteFact(f); err != nil {
		http.Error(rw, "404: Fact not found", http.StatusNotFound)
		return
	}

	c.SetNotificationMessage(rw, req, "Fact deleted!")
	http.Redirect(rw, req.Request, ListFactUrl.Make(), http.StatusFound)
}
