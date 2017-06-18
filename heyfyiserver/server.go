package heyfyiserver

import (
	"encoding/gob"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gocraft/web"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/kiwih/heyfyi/heyfyiserver/fact"
	"github.com/kiwih/heyfyi/heyfyiserver/fyidb"
)

var (
	store *sessions.CookieStore

	templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("./media/templates/*")) //this initializes the template engine
	decoder   = schema.NewDecoder()                                                       //this initializes the schema (HTML form decoding) engine
)

func StartServer(serverAddress string, cookieStoreSalt string) {

	//gob is used when we save failed form structs to the session
	gob.Register(CreateAccount{})
	gob.Register(fact.Fact{})

	decoder.RegisterConverter(false, ConvertBool)

	fyidb.ConnectDatabase("heyfyi")

	store = sessions.NewCookieStore([]byte(cookieStoreSalt))

	router := initRouter()

	go BackgroundVoteGiver()

	log.Println("Server running at " + serverAddress)
	if err := http.ListenAndServe(serverAddress, router); err != nil {
		log.Println("Error:", err.Error())
	}
}

func ConvertBool(value string) reflect.Value {
	if value == "on" {
		return reflect.ValueOf(true)
	} else if v, err := strconv.ParseBool(value); err == nil {
		return reflect.ValueOf(v)
	}

	return reflect.ValueOf(false)
}

func ReturnJSON(rw web.ResponseWriter, object interface{}) {
	j, err := json.MarshalIndent(object, "", "\t")
	if err != nil {
		http.Error(rw, "500: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Write(j)
}

func BackgroundVoteGiver() {
	fyidb.DbStorage.GiveOneVoteToAllAccounts()
	time.Sleep(3600 * time.Second)
	BackgroundVoteGiver()
}
