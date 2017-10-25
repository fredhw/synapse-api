package users

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//MongoStore implements Store for MongoDB
type MongoStore struct {
	session *mgo.Session
	dbname  string
	colname string
}

//NewMongoStore constructs a new MongoStore
func NewMongoStore(sess *mgo.Session, dbName string, collectionName string) *MongoStore {
	if sess == nil {
		panic("nil pointer passed for session")
	}
	return &MongoStore{
		session: sess,
		dbname:  dbName,
		colname: collectionName,
	}
}

//GetByID returns the User with the given ID
func (s *MongoStore) GetByID(id bson.ObjectId) (*User, error) {
	user := &User{}
	col := s.session.DB(s.dbname).C(s.colname)
	err := col.FindId(id).One(user)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

//GetByEmail returns the User with the given email
func (s *MongoStore) GetByEmail(email string) (*User, error) {
	user := &User{}
	col := s.session.DB(s.dbname).C(s.colname)
	err := col.Find(bson.M{"email": email}).Limit(1).One(user)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

//GetByUserName returns the User with the given Username
func (s *MongoStore) GetByUserName(username string) (*User, error) {
	user := &User{}
	col := s.session.DB(s.dbname).C(s.colname)
	err := col.Find(bson.M{"username": username}).Limit(1).One(user)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

//Insert converts the NewUser to a User, inserts
//it into the database, and returns it
func (s *MongoStore) Insert(newUser *NewUser) (*User, error) {
	user, err := newUser.ToUser()
	if err != nil {
		return nil, err
	}
	col := s.session.DB(s.dbname).C(s.colname)
	if err := col.Insert(user); err != nil {
		return nil, err
	}

	return user, nil
}

//Update applies UserUpdates to the given user ID
func (s *MongoStore) Update(userID bson.ObjectId, updates *Updates) error {
	change := mgo.Change{
		Update:    bson.M{"$set": updates}, // $ tells to patch instead of replacing
		ReturnNew: true,                    // returns the old version if not set to true
	}
	user := &User{}
	col := s.session.DB(s.dbname).C(s.colname)
	if _, err := col.FindId(userID).Apply(change, user); err != nil {
		return err
	}
	return nil
}

//Delete deletes the user with the given ID
func (s *MongoStore) Delete(userID bson.ObjectId) error {
	col := s.session.DB(s.dbname).C(s.colname)
	return col.RemoveId(userID)
}
