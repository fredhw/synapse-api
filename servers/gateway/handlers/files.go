package handlers

import (
	"io"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"io/ioutil"

	"github.com/synapse-api/servers/gateway/models/users"
	"github.com/synapse-api/servers/gateway/sessions"
)

// Files struct has
type Files struct {
	FileNames    []string      `json:"fileNames,omitempty"`
	User         *users.User   `json:"user,omitempty"`
}

// FileHandler uploads a file to the server
func (ctx *Context) FileHandler(w http.ResponseWriter, r *http.Request) {

	state := &sessionState{}
	if _, err := sessions.GetState(r, ctx.signingKey, ctx.sessionStore, state); err != nil {
		http.Error(w, fmt.Sprintf("error retrieving session state: %v", err), http.StatusInternalServerError)
		return
	}
	switch r.Method {
	case "GET":
		fmt.Println("fetching files...")

		// check for directory
		path := "/root/gateway/raw-data/" + state.User.UserName

		fmt.Printf("user-path: %s\n", path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, os.ModePerm)
		}

		// get contents of directory
		files, err := ioutil.ReadDir("/root/gateway/raw-data/" + state.User.UserName)
		if err != nil {
			log.Fatal(err)
		}

		ot := Files{}
		ot.User = state.User

		for _, f := range files {
			ot.FileNames = append(ot.FileNames, f.Name())
		}

		respond(w, ot)

	case "POST":
		// parse file

		//fmt.Println("uploading...")
        // file, handle, err := r.FormFile("file")
        // if err != nil {
        //     fmt.Fprintf(w, "%v", err)
        //     return
        // }
		// defer file.Close()
		// fmt.Println("file parsed")
		
		val := r.Header.Get("filename");
		if len(val) == 0 {
			http.Error(w, "no file specified", http.StatusUnauthorized)
			return
		}

		// check for directory
		path := "/root/gateway/raw-data/" + state.User.UserName

		fmt.Printf("user-path: %s\n", path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, os.ModePerm)
		}

		// look for duplicate file
		files, err := ioutil.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}

		var dupeFile string

		for _, f := range files {
			if f.Name() == val {
				dupeFile = f.Name()
			}
		}

		if len(dupeFile) > 0 {
			fmt.Println("dupe found, removing")
			deleteFile(w, dupeFile, path)
		}
		
		// save file
        if err := saveFile(w, r, path, val); err != nil {
			fmt.Fprintf(w, "%v", err)
            return
		}
		respond(w, state.User)
	
	case "DELETE":
		fmt.Println("deleting...")

		val := r.Header.Get("filename")

		if len(val) == 0 {
			http.Error(w, "no file specified", http.StatusUnauthorized)
			return
		}

		files, err := ioutil.ReadDir("/root/gateway/raw-data/" + state.User.UserName)
		if err != nil {
			log.Fatal(err)
		}

		var deleteFileName string

		for _, f := range files {
			if f.Name() == val {
				deleteFileName = f.Name()
			}
		}
		
		// check for directory
		path := "/root/gateway/raw-data/" + state.User.UserName

		fmt.Printf("user-path: %s\n", path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, os.ModePerm)
		}

		deleteFile(w, deleteFileName, path)

		respond(w, state.User)

    default:
		http.Error(w, "method must be GET, POST, PATCH, or DELETE", http.StatusMethodNotAllowed)
		return
	}
}

func saveFile(w http.ResponseWriter, r *http.Request, path string, filename string) error {
	fmt.Printf("saving file: %v\n", filename)

	//file multipart.File, handle *multipart.FileHeader

    // data, err := ioutil.ReadAll(file)
    // if err != nil {
    //     fmt.Fprintf(w, "%v", err)
    //     return err
    // }

    // err = ioutil.WriteFile(path+"/"+handle.Filename, data, 0666)
    // if err != nil {
    //     fmt.Fprintf(w, "%v", err)
    //     return err
	// }

	dec := base64.NewDecoder(base64.StdEncoding, r.Body)

	f, err := os.Create(path+"/"+filename)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, dec); err != nil {
		fmt.Fprintf(w, "%v", err)
		return err
	}
	
	w.WriteHeader(http.StatusCreated)
	return nil
}

func deleteFile(w http.ResponseWriter, deleteFileName string, path string) {
	if len(deleteFileName) == 0 {
		http.Error(w, "file not found", http.StatusUnauthorized)
		return
	}

	fullpath := path + "/" + deleteFileName
	//fmt.Printf("fullpath: %v\n", fullpath)

	if err := os.Remove(fullpath); err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	fmt.Println("deleted")
}