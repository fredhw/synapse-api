package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/mgo.v2"

	"github.com/challenges-fredhw/servers/gateway/models/users"
	"github.com/challenges-fredhw/servers/gateway/sessions"

	"github.com/go-redis/redis"

	"github.com/challenges-fredhw/servers/gateway/handlers"
)

//RootHandler handles requests for the root resource
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprintf(w, "Hello from the gateway! Try requesting /v1/summary/")
}

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

	messageSvcAddrs := os.Getenv("MESSAGESSVC_ADDRS")
	splitMessageSvcAddrs := strings.Split(messageSvcAddrs, ",")
	if len(splitMessageSvcAddrs) == 0 {
		splitMessageSvcAddrs = append(splitMessageSvcAddrs, ":80")
	}

	summarySvcAddrs := os.Getenv("SUMMARYSVC_ADDRS")
	splitSummarySvcAddrs := strings.Split(summarySvcAddrs, ",")
	if len(splitSummarySvcAddrs) == 0 {
		splitSummarySvcAddrs = append(splitSummarySvcAddrs, ":80")
	}

	qeegSvcAddrs := os.Getenv("QEEGSVC_ADDRS")
	splitQeegSvcAddrs := strings.Split(qeegSvcAddrs, ",")
	if len(splitQeegSvcAddrs) == 0 {
		splitQeegSvcAddrs = append(splitQeegSvcAddrs, ":80")
	}

	handlerCtx := handlers.NewHandlerContext(sskey, mongoStore, redisStore)

	mux := http.NewServeMux()
	mux.HandleFunc("/", RootHandler)

	mux.HandleFunc("/v1/users/", handlerCtx.UsersHandler)
	mux.HandleFunc("/v1/users/me/", handlerCtx.UsersMeHandler)
	mux.HandleFunc("/v1/sessions/", handlerCtx.SessionsHandler)
	mux.HandleFunc("/v1/sessions/mine/", handlerCtx.SessionsMineHandler)
	mux.HandleFunc("/v1/users", handlerCtx.SearchHandler)

	mux.Handle("/v1/channels", handlerCtx.NewServiceProxy(splitMessageSvcAddrs))
	mux.Handle("/v1/channels/", handlerCtx.NewServiceProxy(splitMessageSvcAddrs))
	mux.Handle("/v1/messages/", handlerCtx.NewServiceProxy(splitMessageSvcAddrs))
	mux.Handle("/v1/summary/", handlerCtx.NewServiceProxy(splitSummarySvcAddrs))
	mux.Handle("/v1/hello", handlerCtx.NewServiceProxy(splitQeegSvcAddrs))
	mux.Handle("/v1/spectrum/", handlerCtx.NewServiceProxy(splitQeegSvcAddrs))
	mux.Handle("/v1/sumfile/", handlerCtx.NewServiceProxy(splitQeegSvcAddrs))
	mux.Handle("/v1/specfile/", handlerCtx.NewServiceProxy(splitQeegSvcAddrs))
	mux.Handle("/v1/cohrfile/", handlerCtx.NewServiceProxy(splitQeegSvcAddrs))

	corsHandler := handlers.NewCORSHandler(mux)

	fmt.Printf("server is listening on %s\n", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlscert, tlskey, corsHandler))
}
