package repository

// HelloRepository handles simple message storage
type HelloRepository struct {
	message string
}

// NewHelloRepository creates a new hello repository
func NewHelloRepository() *HelloRepository {
	return &HelloRepository{
		message: "Hello World",
	}
}

// GetMessage returns the stored message
func (r *HelloRepository) GetMessage() string {
	return r.message
}