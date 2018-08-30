package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// IsExist returns whehther collection has the query doc.
func IsExist(c *mgo.Collection, query bson.M) (bool, error) {
	cnt, err := c.Find(query).Count()
	if err != nil {
		return false, err
	}
	if cnt > 0 {
		return true, nil
	}
	return false, nil
}
