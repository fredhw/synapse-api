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
