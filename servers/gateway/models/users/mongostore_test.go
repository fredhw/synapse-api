package users

import (
	"testing"

	"gopkg.in/mgo.v2"
)

/*
TestMemStore tests the MemStore object

Since a Store is like a database, you can't really test methods like Get()
or Delete() without also calling (and therefore testing) methods like Save(),
so instead of testing individual methods in isolation, this test runs through
a full CRUD cycle, ensuring the correct behavior occurs at each point in that
cycle. You should use a similar approach when testing your RedisStore implementation.
*/
func TestMongoStore(t *testing.T) {

	nu := NewUser{
		Email:        "foobar@uw.edu",
		Password:     "123456",
		PasswordConf: "123456",
		UserName:     "foobar",
		FirstName:    "Frederick",
		LastName:     "Wijaya",
	}

	upd := &Updates{
		FirstName: "Fred",
		LastName:  "Harrison",
	}

	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		t.Fatalf("error connecting to mongo: %v", err)
	}

	store := NewMongoStore(session, "mgo", "content")

	if _, err := store.GetByEmail(nu.Email); err != ErrUserNotFound {
		t.Errorf("incorrect error when getting user that was never stored: expected %v but got %v", ErrUserNotFound, err)
	}

	if _, err := store.GetByUserName(nu.UserName); err != ErrUserNotFound {
		t.Errorf("incorrect error when getting user that was never stored: expected %v but got %v", ErrUserNotFound, err)
	}

	user, err := store.Insert(&nu)
	if err != nil {
		t.Fatalf("error inserting user: %v", err)
	}

	if err := store.Update(user.ID, upd); err != nil {
		t.Fatalf("error updating user: %v", err)
	}

	user2, err := store.GetByID(user.ID)
	if err != nil {
		t.Fatalf("error getting user from ID: %v", err)
	}

	if _, err := store.GetByUserName(nu.UserName); err != nil {
		t.Fatalf("error getting user from UserName: %v", err)
	}

	if _, err := store.GetByEmail(nu.Email); err != nil {
		t.Fatalf("error getting user from Email:: %v", err)
	}

	if user2.FirstName != upd.FirstName || user2.LastName != upd.LastName {
		t.Errorf("error in updated name: expected %s but got %s", user.FullName(), user2.FullName())
	}

	if err := store.Delete(user.ID); err != nil {
		t.Errorf("error deleting state: %v", err)
	}

	if _, err := store.GetByID(user.ID); err != ErrUserNotFound {
		t.Fatalf("incorrect error when getting state that was deleted: expected %v but got %v", ErrUserNotFound, err)
	}

	if _, err := store.GetByEmail(nu.Email); err != ErrUserNotFound {
		t.Errorf("incorrect error when getting user that was never stored: expected %v but got %v", ErrUserNotFound, err)
	}

	if _, err := store.GetByUserName(nu.UserName); err != ErrUserNotFound {
		t.Errorf("incorrect error when getting user that was never stored: expected %v but got %v", ErrUserNotFound, err)
	}
}
