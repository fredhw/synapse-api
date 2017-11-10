package indexes

import "testing"
import "gopkg.in/mgo.v2/bson"

//TODO: implement automated tests for your trie data structure

func TestTrie(t *testing.T) {
	var origin []bson.ObjectId
	tr := NewTrie()

	id1 := bson.NewObjectId()
	id2 := bson.NewObjectId()
	s := append(origin, id1, id2)

	tr.Add("ac", id2)
	tr.Add("aba", id1) // check for alphabetical ordering
	tr.Add("ac", id2)  // check for adding same key/value pair

	s2 := tr.Get(3, "a")

	if len(s) != len(s2) {
		t.Errorf("error unequal lengths: expected:%d , actual:%d", len(s), len(s2))
	}

	for i := range s {
		if s[i] != s2[i] {
			t.Errorf("error non-matching values: expected %v but got %v", s[i], s2[i])
		}
	}

	id3 := bson.NewObjectId()
	id4 := bson.NewObjectId()
	tr.Add("acb", id4)
	tr.Add("aca", id3)

	s3 := append(origin, id1, id3, id4)
	s4 := tr.Get(3, "a")

	// check for limit handling
	if len(s3) != len(s4) {
		t.Errorf("error unequal lengths: expected:%d , actual:%d", len(s3), len(s4))
	}

	// check for alphabetical sorted depth-first search
	for i := range s3 {
		if s3[i] != s4[i] {
			t.Errorf("error non-matching values: expected %v but got %v", s3[i], s4[i])
		}
	}

	if err := tr.Remove("aba", id1); err != nil {
		t.Errorf("error removing value: expected nil but got %v", err)
	}

	//check for removed node at value
	if res := tr.Get(3, "aba"); len(res) != 0 {
		t.Errorf("expected empty slice but got %v", res)
	}

	//check for removed empty node
	if res := tr.Get(3, "ab"); len(res) != 0 {
		t.Errorf("expected empty slice but got %v", res)
	}
}
