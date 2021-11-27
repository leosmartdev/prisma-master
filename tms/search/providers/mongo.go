// Package provides provide implementations for search interface.
package providers

import (
	"prisma/tms/db/mongo"

	"github.com/globalsign/mgo/bson"
)

// Mongo allows to search data in mongodb
type Mongo struct {
	c *mongo.MongoClient
}

// NewMongoSearchProvider returns an instance of mongo search structure
func NewMongoSearchProvider(c *mongo.MongoClient) *Mongo {
	return &Mongo{
		c: c,
	}
}

func (m *Mongo) Search(text string, fields bson.M, collections []string, limit int) (records []interface{}, err error) {
	ret := make([]interface{}, 0)
	var query bson.D
	if text != "" {
		query = append(query, bson.DocElem{
			Name: "$text",
			Value: bson.M{
				"$search": text,
			},
		})
	}
	for name, value := range fields {
		query = append(query, bson.DocElem{
			Name: name,
			Value: value,
		})
	}
	for _, collection := range collections {
		err := m.c.DB().C(collection).Find(query).Limit(limit).All(&ret)
		m.c.Sess().Close()
		return ret, err
	}
	return ret, nil
}
