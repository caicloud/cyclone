package models

import (
	"reflect"

	"github.com/caicloud/nirvana/log"

	"gopkg.in/mgo.v2"
)

func sameOptions(idx1, idx2 mgo.Index) bool {
	return idx1.Unique == idx2.Unique &&
		idx1.DropDups == idx2.DropDups &&
		idx1.Background == idx2.Background &&
		idx1.Sparse == idx2.Sparse &&
		idx1.ExpireAfter == idx2.ExpireAfter &&
		idx1.Min == idx2.Min &&
		idx1.Max == idx2.Max &&
		idx1.Minf == idx2.Minf &&
		idx1.Maxf == idx2.Maxf &&
		idx1.BucketSize == idx2.BucketSize &&
		idx1.Bits == idx2.Bits &&
		idx1.DefaultLanguage == idx2.DefaultLanguage &&
		idx1.LanguageOverride == idx2.LanguageOverride &&
		reflect.DeepEqual(idx1.Weights, idx2.Weights) &&
		reflect.DeepEqual(idx1.Collation, idx2.Collation)
}

func existAndChanged(indexes []mgo.Index, index mgo.Index) bool {
	for _, idx := range indexes {
		if reflect.DeepEqual(idx.Key, index.Key) && !sameOptions(idx, index) {
			return true
		}
	}
	return false
}

func EnsureIndexes(c *mgo.Collection, idxes []mgo.Index) {
	indexes, _ := c.Indexes()
	for _, idx := range idxes {
		if existAndChanged(indexes, idx) {
			log.Infof("Drop existing index %v for collection %s since options changed", idx.Key, c.Name)
			if err := c.DropIndex(idx.Key...); err != nil {
				log.Fatalf("Drop existing index error: %v", err)
			}
		}

		err := c.EnsureIndex(idx)
		if err != nil {
			log.Fatal(err)
		}
	}
}
