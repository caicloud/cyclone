package cyclone_test

import (
	. "github.com/onsi/gomega"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"fmt"
)

const (
	defaultDBName string = "cyclone"
)

type Mongo interface {
	Insert(string, interface{}) error
	Update(string, string, interface{}) error
	Find(string, string, interface{}) error
	Remove(string, string) error
	Close()
}

type mongo struct {
	s *mgo.Session
	c map[string]*mgo.Collection
}

// NewMongoClient return a mongo client setting up mongo items in test
func NewMongoClient(host string, collections []string) Mongo {
	session, err := mgo.Dial(host)
	Expect(err).NotTo(HaveOccurred())

	session.SetMode(mgo.Eventual, true)

	m := new(mongo)

	m.s = session

	m.c = make(map[string]*mgo.Collection)
	for _, c := range collections {
		m.c[c] = session.DB(defaultDBName).C(c)
	}

	return m
}

// Insert inserts item
func (m *mongo) Insert(collection string, item interface{}) error {
	if _, ok := m.c[collection]; !ok {
		return fmt.Errorf("no collections named %s", collection)
	}

	if err := m.c[collection].Insert(item); err != nil {
		return err
	}

	return nil
}

// Find finds item named $name and put it in result
func (m *mongo) Find(collection, name string, result interface{}) error {
	if _, ok := m.c[collection]; !ok {
		return fmt.Errorf("no collections named %s", collection)
	}

	query := bson.M{"name": name}
	count, err := m.c[collection].Find(query).Count()
	if err != nil {
		return err
	}

	if count == 0 {
		return mgo.ErrNotFound
	} else if count > 1 {
		return fmt.Errorf("there are %d items with the same name %s", count, name)
	}

	if err = m.c[collection].Find(query).One(result); err != nil {
		return err
	}

	return nil
}

// Update updates the item named $name
func (m *mongo) Update(collection, name string, item interface{}) error {
	if _, ok := m.c[collection]; !ok {
		return fmt.Errorf("no collections named %s", collection)
	}

	query := bson.M{"name": name}

	count, err := m.c[collection].Find(query).Count()
	if err != nil {
		return err
	}

	if count == 0 {
		return mgo.ErrNotFound
	} else if count > 1 {
		return fmt.Errorf("there are %d items with the same name %s", count, name)
	}

	return m.c[collection].Update(query, item)
}

// Remove removes item named $name
// Note that if $name == "", Remove will remove all items in the collection
func (m *mongo) clear() error {
	var err error
	for c, _ := range m.c {
		err = m.Remove(c, "")
		if err != nil {
			return nil
		}
	}
	return nil
}

func (m *mongo) Remove(collection, name string) error {
	if name == "" {
		_, err := m.c[collection].RemoveAll(nil)
		return err
	}

	if _, ok := m.c[collection]; !ok {
		return nil
	}

	query := bson.M{"name": name}
	return m.c[collection].Remove(query)
}

func (m *mongo) Close() {
	err := m.clear()
	Expect(err).NotTo(HaveOccurred())

	m.s.Close()
}
