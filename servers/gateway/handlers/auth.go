package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/challenges-fredhw/servers/gateway/models/users"
	"github.com/challenges-fredhw/servers/gateway/sessions"
)

//TODO: define HTTP handler functions as described in the
//assignment description. Remember to use your handler context
//struct as the receiver on these functions so that you have
//access to things like the session store and user store.

//UsersHandler function handles requests for the "users" resource,
//and allows clients to create new user accounts. The method must be POST
//and the request body must contain JSON that can be decoded into a
//users.NewUser struct.
func (ctx *Context) UsersHandler(w http.ResponseWriter, r *http.Request) {
	//If the method is POST, follow these steps:
	//Decode the request body into a users.NewUser struct
	//Validate the NewUser
	//Ensure there isn't already a user in the user store with the same email address
	//Ensure there isn't already a user in the user store with the same user name
	//Insert the new user into the user store
	//Begin a new session
	//Respond to the client with an http.StatusCreated status code,
	//and the users.User struct returned from the user store insert method encoded as a JSON object
	switch r.Method {
	case "POST":
		nu := users.NewUser{}
		if err := json.NewDecoder(r.Body).Decode(&nu); err != nil {
			http.Error(w, fmt.Sprintf("error decoding JSON: %v", err), http.StatusBadRequest)
			return
		}
		if err := nu.Validate(); err != nil {
			http.Error(w, fmt.Sprintf("error validating user: %v", err), http.StatusInternalServerError)
			return
		}

		if _, err := ctx.userStore.GetByEmail(nu.Email); err != users.ErrUserNotFound {
			http.Error(w, fmt.Sprintf("there is already a user with same email address: %v", err), http.StatusBadRequest)
			return
		}

		if _, err := ctx.userStore.GetByUserName(nu.UserName); err != users.ErrUserNotFound {
			http.Error(w, fmt.Sprintf("there is already a user with same user name: %v", err), http.StatusBadRequest)
			return
		}

		user, err := ctx.userStore.Insert(&nu)
		if err != nil {
			http.Error(w, fmt.Sprintf("error inserting user: %v", err), http.StatusInternalServerError)
			return
		}

		state := &sessionState{
			Time: time.Now(),
			User: user,
		}

		if _, err := sessions.BeginSession(ctx.signingKey, ctx.sessionStore, state, w); err != nil {
			http.Error(w, "error beginning session", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		respond(w, user)

	default:
		http.Error(w, "method must be POST", http.StatusMethodNotAllowed)
		return
	}
}

//UsersMeHandler handles requests for the "current user" resource
func (ctx *Context) UsersMeHandler(w http.ResponseWriter, r *http.Request) {
	//GET: get the current user from the session state and respond with that user encoded as JSON object.
	//PATCH: update the current user with the JSON in the request body, and respond with the newly updated user,
	//encoded as a JSON object. Remember that you are also caching the current user data in your session store, so update that as well.
	switch r.Method {
	case "GET":
		state := &sessionState{}
		if _, err := sessions.GetState(r, ctx.signingKey, ctx.sessionStore, state); err != nil {
			http.Error(w, fmt.Sprintf("error retrieving session state: %v", err), http.StatusInternalServerError)
			return
		}

		respond(w, state.User)

	case "PATCH":
		//get state from context
		state := &sessionState{}
		sid, err := sessions.GetState(r, ctx.signingKey, ctx.sessionStore, state)
		if err != nil {
			http.Error(w, fmt.Sprintf("error retrieving session state: %v", err), http.StatusInternalServerError)
			return
		}

		//retrieve updates
		upd := users.Updates{}
		if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
			http.Error(w, fmt.Sprintf("error decoding JSON: %v", err), http.StatusBadRequest)
			return
		}

		//apply updates
		if err := state.User.ApplyUpdates(&upd); err != nil {
			http.Error(w, fmt.Sprintf("error updating user: %v", err), http.StatusBadRequest)
			return
		}
		if err := ctx.sessionStore.Save(sid, state); err != nil {
			http.Error(w, fmt.Sprintf("error updating user in store: %v", err), http.StatusBadRequest)
			return
		}

		respond(w, state.User)
	default:
		http.Error(w, "method must be GET or PATCH", http.StatusMethodNotAllowed)
		return
	}
}

//SessionsHandler handles requests for the "sessions" resource, and allows clients to begin a new session
//using an existing user's credentials. The method must be POST and the request body must contain JSON that
//can be decoded into a users.Credentials struct.
func (ctx *Context) SessionsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		cd := users.Credentials{}
		if err := json.NewDecoder(r.Body).Decode(&cd); err != nil {
			http.Error(w, fmt.Sprintf("error decoding JSON: %v", err), http.StatusBadRequest)
			return
		}

		user, err := ctx.userStore.GetByEmail(cd.Email)
		if err != nil {
			http.Error(w, "invalid email", http.StatusUnauthorized)
			return
		}

		if err := user.Authenticate(cd.Password); err != nil {
			http.Error(w, "invalid password", http.StatusUnauthorized)
			return
		}

		state := &sessionState{
			Time: time.Now(),
			User: user,
		}

		if _, err := sessions.BeginSession(ctx.signingKey, ctx.sessionStore, state, w); err != nil {
			http.Error(w, "error beginning session", http.StatusInternalServerError)
			return
		}

		respond(w, user)
	default:
		http.Error(w, "method must be POST", http.StatusMethodNotAllowed)
		return
	}
}

//SessionsMineHandler handles requests for the "current session" resource, and allows clients to end that session.
func (ctx *Context) SessionsMineHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		if _, err := sessions.EndSession(r, ctx.signingKey, ctx.sessionStore); err != nil {
			http.Error(w, fmt.Sprintf("error ending session: %v", err), http.StatusInternalServerError)
		}

		w.Header().Add(headerContentType, "text/plain")
		fmt.Fprintln(w, "signed out")

	default:
		http.Error(w, "method must be DELETE", http.StatusMethodNotAllowed)
		return
	}
}

//respond encodes `value` into JSON and writes that to the response
func respond(w http.ResponseWriter, value interface{}) {
	w.Header().Add(headerContentType, contentTypeJSON)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, fmt.Sprintf("error encoding response value to JSON: %v", err), http.StatusInternalServerError)
	}
}
