package indexes

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"gopkg.in/mgo.v2/bson"
)

//TODO: implement a trie data structure that stores
//keys of type string and values of type bson.ObjectId
//It must store key/value pairs, where the key is a string and the value is a bson.ObjectId.
//Multiple values may be stored for the same key, but they should remain a distinct set.
//That is, if the key/value is already in the trie, don't add the value again.

//It must have a method that returns the first n values that match a given prefix string,
//where n and the prefix string are parameters.
//Find the branch of the trie holding keys that start with the prefix string,
//and then do a recursive depth-first search to find the first n values in that branch.

//It must have a method that removes a given key/value pair from the trie.
//Since this trie will be read and modified from multiple concurrent goroutines,
//you must protect the node tree using a RWMutex.

//Node stores a value
type Node struct {
	value []bson.ObjectId
	child map[string]*Node
	depth int
}

//Trie is a TTL cache that is safe for concurrent use
type Trie struct {
	root *Node
	mx   sync.RWMutex
}

//NewTrie constructs a new Trie object
func NewTrie() *Trie {
	return &Trie{
		root: &Node{
			child: make(map[string]*Node),
			value: [](bson.ObjectId){},
			depth: 0,
		},
	}
}

//NewChild constructs a new node corresponding to the given key
func (n *Node) NewChild(key string) *Node {
	node := &Node{
		value: [](bson.ObjectId){},
		child: make(map[string]*Node),
		depth: n.depth + 1,
	}
	n.child[key] = node
	return node
}

//let current node = root node
//for each letter in the key...
//	find the child node of current node associated with that letter
//	if there is no child node associated with that letter, create a new node and add it to current node as a child associated with the letter
//	set current node = child node
//add value to current node

//Add puts the id into the Trie
func (c *Trie) Add(name string, value bson.ObjectId) {
	c.mx.Lock()
	defer c.mx.Unlock()
	current := c.root
	letters := strings.Split(name, "")
	for i := range letters {
		next, found := current.child[letters[i]]
		if !found {
			next = current.NewChild(letters[i])
		}
		current = next
	}
	for _, stored := range current.value {
		if stored == value {
			return
		}
	}
	current.value = append(current.value, value)
}

//ToString returns a string representation of the key/value pair
func (c *Trie) ToString(name string) string {
	c.mx.Lock()
	defer c.mx.Unlock()
	result := "result:"
	current := c.root
	letters := strings.Split(name, "")
	for _, k := range letters {
		next, found := current.child[k]
		if !found {
			return "key not found"
		}
		result = fmt.Sprintf("%s [%s]%d", result, k, next.depth)
		current = next
	}
	return result
}

//Remove removes the key/value pair from the Trie
func (c *Trie) Remove(key string, value bson.ObjectId) error {
	c.mx.Lock()
	defer c.mx.Unlock()
	current := c.root
	letters := strings.Split(key, "")

	return current.RemoveHelper(letters, value)
}

//RemoveHelper removes empty child nodes after value is removed
func (n *Node) RemoveHelper(keys []string, value bson.ObjectId) error {
	if len(keys) == 0 {
		return nil
	}
	next, found := n.child[keys[0]]
	if !found {
		return fmt.Errorf("error no key value at depth %v, key %v", n.depth, keys)
	}

	for i, v := range next.value {
		if v == value {
			next.value = append(next.value[:i], next.value[i+1:]...)
			break
		}
	}

	if err := next.RemoveHelper(keys[1:], value); err != nil {
		return err
	}

	if len(next.child) == 0 && len(next.value) == 0 {
		delete(n.child, keys[0])
	}
	return nil
}

//Get returns the first n values that match a given prefix string,
//where n and the prefix string are parameters.
//Find the branch of the trie holding keys that start with the prefix string,
//and then do a recursive depth-first search to find the first n values in that branch.
func (c *Trie) Get(n int, prefix string) []bson.ObjectId {
	c.mx.RLock()
	defer c.mx.RUnlock()
	vals := []bson.ObjectId{}
	if len(prefix) == 0 {
		return vals
	}
	current := c.root
	letters := strings.Split(prefix, "")

	for i := range letters {
		next, found := current.child[letters[i]]
		if !found {
			return vals
		}
		current = next
	}

	// do recursion
	res := current.GetHelper(n)
	for _, v := range res {
		vals = append(vals, v)
	}

	return vals
}

//GetHelper does a recursive depth-first search
func (n *Node) GetHelper(limit int) []bson.ObjectId {
	vals := []bson.ObjectId{}
	lim := limit
	keys := sortedKeys(n.child)
	for _, key := range keys {
		// append all values from the deeper searches to vals
		next := n.child[key]
		arr := next.GetHelper(lim)
		for _, v := range arr {
			if lim == 0 {
				return vals
			}
			vals = append(vals, v)
			lim = lim - 1
		}
	}

	// append the value (if any) from current node
	if len(n.value) > 0 {
		for _, v := range n.value {
			if lim == 0 {
				return vals
			}
			vals = append(vals, v)
			lim = lim - 1
		}
	}
	return vals
}

func sortedKeys(m map[string]*Node) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
