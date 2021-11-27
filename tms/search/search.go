// Package search provides interface to search any information.
package search

import "github.com/globalsign/mgo/bson"

// Searcher is an interface for passing different search providers
type Searcher interface {
	// Search records in a document or a table(which are pointed out in where) of a database
	Search(text string, fields bson.M, where []string, limit int) (records []interface{}, err error)
}
