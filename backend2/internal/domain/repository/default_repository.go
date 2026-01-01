package repository

type DefaultRepository interface {
	// Create inserts a new record into the database.
	Create(entity *interface{}) error
	// FindByID retrieves a record by its ID.
	FindByID(id uint64) (*interface{}, error)
	// FindBy any field retrieves a record by a specific field and value.
	FindBy(field string, value any) (*interface{}, error)
	// List retrieves a list of records with pagination.
	List(page, size int) ([]interface{}, int64, error)
	// Update modifies an existing record.
	Update(entity *interface{}) error
	// Delete removes a record by its ID.
	Delete(id uint64) error
	// Count returns the total number of records for a given entity type.
	Count() (int64, error)
	// GetAll retrieves all records of a specific type.
	GetAll() ([]interface{}, error)
	// FindBy any field retrieves a record by a specific field and value.
	GetBy(interface{}) ([]interface{}, error)
}
