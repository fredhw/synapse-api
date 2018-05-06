package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"github.com/synapse-api/servers/gateway/indexes"
	"github.com/synapse-api/servers/gateway/models/users"
	"github.com/synapse-api/servers/gateway/sessions"
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

		addToTrie(user, ctx.trie)

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
		if err := ctx.userStore.Update(state.User.ID, &upd); err != nil {
			http.Error(w, fmt.Sprintf("error updating user: %v", err), http.StatusBadRequest)
			return
		}

		if err := addAndRemove(state.User, ctx.trie, &upd); err != nil {
			http.Error(w, fmt.Sprintf("error with trie: %v", err), http.StatusInternalServerError)
			return
		}

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

//SearchHandler handles user search requests for authenticated users
func (ctx *Context) SearchHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//get state from context
		state := &sessionState{}
		_, err := sessions.GetState(r, ctx.signingKey, ctx.sessionStore, state)
		if err != nil {
			http.Error(w, fmt.Sprintf("error retrieving session state: %v", err), http.StatusUnauthorized)
			return
		}

		q := r.URL.Query().Get("q")
		if len(q) == 0 {
			respond(w, "")
		}
		q = strings.ToLower(q)
		ids := ctx.trie.Get(20, q)
		users := ctx.userStore.GetByIDSlice(ids)
		respond(w, users)

	default:
		http.Error(w, "method must be GET", http.StatusMethodNotAllowed)
		return
	}
}

//NewServiceProxy uses addresses to create reverse proxies for microservices
func (ctx *Context) NewServiceProxy(addrs []string) *httputil.ReverseProxy {
	nextIndex := 0
	mx := sync.Mutex{}
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			state := &sessionState{}
			sessions.GetState(r, ctx.signingKey, ctx.sessionStore, state)
			if state.User != nil {
				userJSON, err := json.Marshal(state.User)
				if err != nil {
					log.Printf("error marshaling user: %v", err)
				}
				r.Header.Add("X-User", string(userJSON))
			} else {
				r.Header.Del("X-User")
			}

			mx.Lock()
			r.URL.Host = addrs[nextIndex%len(addrs)]
			nextIndex++
			mx.Unlock()
			r.URL.Scheme = "http"
		},
	}
}

//respond encodes `value` into JSON and writes that to the response
func respond(w http.ResponseWriter, value interface{}) {
	w.Header().Add(headerContentType, contentTypeJSON)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, fmt.Sprintf("error encoding response value to JSON: %v", err), http.StatusInternalServerError)
	}
}

//addToTrie indexes user fields into the Trie
func addToTrie(user *users.User, trie *indexes.Trie) {
	em := strings.ToLower(user.Email)
	un := strings.ToLower(user.UserName)
	fn := strings.ToLower(user.FirstName)
	ln := strings.ToLower(user.LastName)

	trie.Add(em, user.ID)
	trie.Add(un, user.ID)
	trie.Add(fn, user.ID)
	trie.Add(ln, user.ID)
}

//removeFromTrie removes indexed user fields from the Trie
func addAndRemove(user *users.User, trie *indexes.Trie, upd *users.Updates) error {
	of := strings.ToLower(user.FirstName)
	ol := strings.ToLower(user.LastName)
	fn := strings.ToLower(upd.FirstName)
	ln := strings.ToLower(upd.LastName)

	if err := trie.Remove(of, user.ID); err != nil {
		return fmt.Errorf("error removing first name from trie: %v", err)
	}
	if err := trie.Remove(ol, user.ID); err != nil {
		return fmt.Errorf("error removing lastname from trie: %v", err)
	}
	trie.Add(fn, user.ID)
	trie.Add(ln, user.ID)
	return nil
}


// // Files struct has
// type Files struct {
// 	FileNames    []string      `json:"fileNames,omitempty"`
// 	User         *users.User   `json:"user,omitempty"`
// }

// // FileHandler uploads a file to the server
// func (ctx *Context) FileHandler(w http.ResponseWriter, r *http.Request) {

// 	state := &sessionState{}
// 	if _, err := sessions.GetState(r, ctx.signingKey, ctx.sessionStore, state); err != nil {
// 		http.Error(w, fmt.Sprintf("error retrieving session state: %v", err), http.StatusInternalServerError)
// 		return
// 	}
// 	switch r.Method {
// 	case "GET":
// 		fmt.Println("fetching files...")
// 		files, err := ioutil.ReadDir("/root/gateway/raw-data")
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		ot := Files{}
// 		ot.User = state.User

// 		for _, f := range files {
// 			ot.FileNames = append(ot.FileNames, f.Name())
// 		}

// 		respond(w, ot)

//     case "POST":
//         fmt.Println("uploading...")

//         file, handle, err := r.FormFile("file")
//         if err != nil {
//             fmt.Fprintf(w, "%v", err)
//             return
//         }
//         defer file.Close()
		
// 		fmt.Println("file parsed")

// 		mimeType := handle.Header.Get("Content-Type")
// 		fmt.Printf("checking filetype: %v\n", mimeType)

// 		files, err := ioutil.ReadDir("/root/gateway/raw-data")
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		var dupeFile string

// 		for _, f := range files {
// 			if f.Name() == handle.Filename {
// 				dupeFile = f.Name()
// 			}
// 		}

// 		if len(dupeFile) > 0 {
// 			fmt.Println("dupe found, removing")
// 			deleteFile(w, dupeFile)
// 		}

//         saveFile(w, file, handle)
		
//         w.WriteHeader(http.StatusCreated)
// 		respond(w, state.User)
	
// 	case "DELETE":
// 		fmt.Println("deleting...")

// 		val := r.Header.Get("filename")

// 		if len(val) == 0 {
// 			http.Error(w, "no file specified", http.StatusUnauthorized)
// 			return
// 		}

// 		files, err := ioutil.ReadDir("/root/gateway/raw-data")
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		var deleteFileName string

// 		for _, f := range files {
// 			if f.Name() == val {
// 				deleteFileName = f.Name()
// 			}
// 		}
		
// 		deleteFile(w, deleteFileName)

// 		respond(w, state.User)

//     default:
// 		http.Error(w, "method must be GET, POST, PATCH, or DELETE", http.StatusMethodNotAllowed)
// 		return
// 	}
// }

// func saveFile(w http.ResponseWriter, file multipart.File, handle *multipart.FileHeader) {
// 	fmt.Printf("saving file: %v\n", handle.Filename)

//     data, err := ioutil.ReadAll(file)
//     if err != nil {
//         fmt.Fprintf(w, "%v", err)
//         return
//     }

//     err = ioutil.WriteFile("/root/gateway/raw-data/"+handle.Filename, data, 0666)
//     if err != nil {
//         fmt.Fprintf(w, "%v", err)
//         return
//     }
// }

// func deleteFile(w http.ResponseWriter, deleteFileName string) {
// 	if len(deleteFileName) == 0 {
// 		http.Error(w, "file not found", http.StatusUnauthorized)
// 		return
// 	}

// 	fullpath := fmt.Sprintf("/root/gateway/raw-data/%s", deleteFileName)
// 	//fmt.Printf("fullpath: %v\n", fullpath)

// 	if err := os.Remove(fullpath); err != nil {
// 		fmt.Fprintf(w, "%v", err)
// 		return
// 	}

// 	fmt.Println("deleted")
// }
