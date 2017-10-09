package main

import (
	"log"
	"net/http"
	"os"

	"github.com/challenges-fredhw/servers/gateway/handlers"
)

//main is the main entry point for the server
func main() {
	/* TODO: add code to do the following
	- Read the ADDR environment variable to get the address
	  the server should listen on. If empty, default to ":80"
	- Create a new mux for the web server.
	- Tell the mux to call your handlers.SummaryHandler function
	  when the "/v1/summary" URL path is requested.
	- Start a web server listening on the address you read from
	  the environment variable, using the mux you created as
	  the root handler. Use log.Fatal() to report any errors
	  that occur when trying to start the web server.
	*/

	port := "localhost:80"

	if env := os.Getenv("ADDR"); len(env) > 0 {
		port = env
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/summary/", handlers.SummaryHandler)
	log.Fatal(http.ListenAndServe(port, mux))
}
