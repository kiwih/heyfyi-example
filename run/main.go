package main

import (
	"flag"

	"github.com/kiwih/heyfyi"
)

var (
	logFileName   = flag.String("logfilename", "heyfyi.txt", "Specify the name of the log file")
	serverAddress = flag.String("serveraddress", ":3000", "Specify the address the server should listen on")
)

func main() {
	heyfyi.StartServer(*logFileName, *serverAddress)
}
