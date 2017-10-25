package users

import (
	"encoding/json"
	"time"

	"github.com/patrickmn/go-cache"
	"gopkg.in/mgo.v2/bson"
)

//MemStore represents an in-process memory session store.
//This should be used only for testing and prototyping.
//Production systems should use a shared server store like redis
type MemStore struct {
	entries *cache.Cache
}

//NewMemStore constructs and returns a new MemStore
func NewMemStore(sessionDuration time.Duration, purgeInterval time.Duration) *MemStore {
	return &MemStore{
		entries: cache.New(sessionDuration, purgeInterval),
	}
}

//GetByID returns the User with the given ID
func (ms *MemStore) GetByID(id bson.ObjectId) (*User, error) {
	user := &User{}
	j, found := ms.entries.Get(string(id))
	if !found {
		return nil, ErrUserNotFound
	}

	if err := json.Unmarshal(j.([]byte), user); err != nil {
		return nil, err
	}
	return user, nil
}

//GetByEmail returns the User with the given email
func (ms *MemStore) GetByEmail(email string) (*User, error) {
	user := &User{}
	m := ms.entries.Items()
	for _, v := range m {
		if err := json.Unmarshal((v.Object).([]byte), user); err != nil {
			return nil, err
		}
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

//GetByUserName returns the User with the given Username
func (ms *MemStore) GetByUserName(username string) (*User, error) {
	user := &User{}
	m := ms.entries.Items()
	for _, v := range m {
		if err := json.Unmarshal((v.Object).([]byte), user); err != nil {
			return nil, err
		}
		if user.UserName == username {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

//Insert converts the NewUser to a User, inserts
//it into the database, and returns it
func (ms *MemStore) Insert(newUser *NewUser) (*User, error) {
	user, err := newUser.ToUser()
	if err != nil {
		return nil, err
	}
	j, err := json.Marshal(user)
	if nil != err {
		return nil, err
	}
	ms.entries.Add(string(user.ID), j, cache.DefaultExpiration)
	return user, nil
}

//Update applies UserUpdates to the given user ID
func (ms *MemStore) Update(userID bson.ObjectId, updates *Updates) error {
	user, err := ms.GetByID(userID)
	if err != nil {
		return err
	}
	if err := user.ApplyUpdates(updates); err != nil {
		return err
	}

	j, err := json.Marshal(user)
	if nil != err {
		return err
	}
	ms.entries.Set(string(userID), j, cache.DefaultExpiration)
	return nil
}

//Delete deletes the user with the given ID
func (ms *MemStore) Delete(userID bson.ObjectId) error {
	ms.entries.Delete(string(userID))
	return nil
}
