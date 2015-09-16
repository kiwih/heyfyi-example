package main

import (
	"io"
	"log"
	"os"

	"github.com/kiwih/heyfyi"
)

func main() {
	//Enable logger
	logFileName := os.Getenv("LOG_FILE_NAME")
	if len(logFileName) == 0 {
		logFileName = "heyfyi.txt"
	}

	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("Can't open log file: " + err.Error())
	}
	log.SetOutput(io.MultiWriter(f, os.Stdout))

	// On platforms like heroky/dokku this should be PORT, not HTTP_PORT. On Azure this should be HTTP_PLATFORM_PORT
	serverAddress := ":" + os.Getenv("HTTP_PORT")
	if serverAddress == ":" {
		log.Println("$HTTP_PORT was not set, defaulting to 3000")
		serverAddress = ":3000"
	}

	cookieStoreSalt := os.Getenv("COOKIE_STORE_SALT")
	if len(cookieStoreSalt) == 0 {
		log.Fatal("$COOKIE_STORE_SALT was not set, cannot go on. Exiting.")
	}

	heyfyi.StartServer(serverAddress, cookieStoreSalt)
}
