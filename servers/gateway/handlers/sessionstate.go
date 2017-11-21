package handlers

import (
	"time"

	"github.com/challenges-fredhw/servers/gateway/models/users"
)

//TODO: define a session state struct for this web server
//see the assignment description for the fields you should include
//remember that other packages can only see exported fields!

type sessionState struct {
	Time time.Time
	User *users.User
}
