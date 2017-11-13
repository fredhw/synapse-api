package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/mgo.v2"

	"github.com/challenges-fredhw/servers/gateway/models/users"
	"github.com/challenges-fredhw/servers/gateway/sessions"

	"github.com/go-redis/redis"

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

	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":443"
	}

	tlskey := os.Getenv("TLSKEY")
	tlscert := os.Getenv("TLSCERT")
	if len(tlskey) == 0 || len(tlscert) == 0 {
		log.Fatal("please set TLSKEY and TLSCERT")
	}

	sskey := os.Getenv("SESSIONKEY")
	if len(sskey) == 0 {
		log.Fatal("please set SESSIONKEY")
	}

	redisAddr := os.Getenv("REDISADDR")
	if len(redisAddr) == 0 {
		redisAddr = "redis:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	redisStore := sessions.NewRedisStore(client, 0)

	dbAddr := os.Getenv("DBADDR")
	if len(dbAddr) == 0 {
		dbAddr = "mymongo:27017"
	}

	fmt.Printf("dialing mogo with: %s\n", dbAddr)

	sess, err := mgo.Dial(dbAddr)
	if err != nil {
		log.Fatalf("failed to dial mongodb: %v", err)
	}
	mongoStore := users.NewMongoStore(sess, "mgo", "users")

	handlerCtx := handlers.NewHandlerContext(sskey, mongoStore, redisStore)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/summary/", handlers.SummaryHandler)
	mux.HandleFunc("/v1/users/", handlerCtx.UsersHandler)
	mux.HandleFunc("/v1/users/me/", handlerCtx.UsersMeHandler)
	mux.HandleFunc("/v1/sessions/", handlerCtx.SessionsHandler)
	mux.HandleFunc("/v1/sessions/mine/", handlerCtx.SessionsMineHandler)
	mux.HandleFunc("/v1/users", handlerCtx.SearchHandler)

	corsHandler := handlers.NewCORSHandler(mux)

	fmt.Printf("server is listening on %s\n", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlscert, tlskey, corsHandler))
}

// //UsersHandler lol
// func UsersHandler(mongosess *mgo.Session, ctx *handlers.Context) func(http.ResponseWriter, *http.Request) {
// 	if mongosess == nil {
// 		panic("nil MongoDB session")
// 	}
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		ctx.UsersHandler(w, r)
// 	}
// }
