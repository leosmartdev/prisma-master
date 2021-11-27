package db

// Migrating is used to migrate db
type Migrating interface {
	// EnsureSetUp executes schema files that are separated by ','
	EnsureSetUp(schemas []string)
}
