package load

// ModelCreator is an interface for a benchmark to do the initial setup of a database
// in preparation for running a benchmark against it.
type ModelCreator interface {
	// Init should set up any connection or other setup for talking to the DB, but should NOT create any databases
	Init()

	// ModelExists checks if a database with the given name currently exists.
	ModelExists(modelName string) bool

	// CreateModel creates a database with the given name.
	CreateModel(modelName string) error

	// RemoveOldModel removes an existing database with the given name.
	RemoveOldModel(modelName string) error
}

// ModelCreatorCloser is a ModelCreator that also needs a Close method to cleanup any connections
// after the benchmark is finished.
type ModelCreatorCloser interface {
	ModelCreator

	// Close cleans up any database connections
	Close()
}

// ModelCreatorPost is a ModelCreator that also needs to do some initialization after the
// database is created (e.g., only one client should actually create the DB, so
// non-creator clients should still set themselves up for writing)
type ModelCreatorPost interface {
	ModelCreator

	// PostCreateModel does further initialization after the database is created
	PostCreateModel(modelName string) error
}
