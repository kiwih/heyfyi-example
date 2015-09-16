package heyfyi

import (
	"encoding/gob"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/gocraft/web"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/kiwih/heyfyi/fact"
	"github.com/kiwih/heyfyi/fyidb"
)

var (
	COOKIE_STORE_SALT string		/* Cryptographic salt for cookie signing/encrypting */
	HTTP_PORT string						/* TCP port for web server */

	store sessions.CookieStore

	templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("./templates/*"))                 //this initializes the template engine
	decoder   = schema.NewDecoder()                                                                       //this initializes the schema (HTML form decoding) engine
)

func StartServer(logFileName string, serverAddress string) {

	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("Can't open log file: " + err.Error())
	}
	log.SetOutput(io.MultiWriter(f, os.Stdout))


	/* On platforms like heroky/dokku this should be PORT, not HTTP_PORT. */
	HTTP_PORT = os.Getenv("HTTP_PORT")
	if len(HTTP_PORT) == 0 {
		log.Println("$HTTP_PORT was not set, defaulting to 3000")
		HTTP_PORT = "3000"
	}

	COOKIE_STORE_SALT = os.Getenv("COOKIE_STORE_SALT")
	if len(COOKIE_STORE_SALT) == 0 {
		log.Fatal("$COOKIE_STORE_SALT was not set, cannot go on. Exiting.")
	} else {
		store = sessions.NewCookieStore([]byte("9s7YD807h*&DHhihSD123434SASDD__834HUSJNCxczc123!@#sd85"))
	}



	//gob is used when we save failed form structs to the session
	gob.Register(CreateAccount{})
	gob.Register(fact.Fact{})

	decoder.RegisterConverter(false, ConvertBool)

	fyidb.ConnectDatabase("heyfyi")

	router := initRouter()

	go BackgroundVoteGiver()

	log.Println("Server running at " + serverAddress)
	if err := http.ListenAndServe(":"+HTTP_PORT, router); err != nil {
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
