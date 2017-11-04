package handlers

import (
	"github.com/challenges-fredhw/servers/gateway/models/users"
	"github.com/challenges-fredhw/servers/gateway/sessions"
)

//TODO: define a handler context struct that
//will be a receiver on any of your HTTP
//handler functions that need access to
//globals, such as the key used for signing
//and verifying SessionIDs, the session store
//and the user store

//Context holds context values used by multiple handler functions.
type Context struct {
	signingKey   string
	userStore    users.Store
	sessionStore sessions.Store
}

//NewHandlerContext returns a struct that
//will be a receiver on any of your HTTP
//handler functions that need access to
//globals, such as the key used for signing
//and verifying SessionIDs, the session store
//and the user store
func NewHandlerContext(key string, userStore users.Store, sessionStore sessions.Store) *Context {
	return &Context{
		signingKey:   key,
		userStore:    userStore,
		sessionStore: sessionStore,
	}
}
