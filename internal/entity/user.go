package entity

// User represents a user.
type User struct {
	ID        string
	Name      string
	Admin     bool
	Resources []string
}

// GetID returns the user ID.
func (u User) GetID() string {
	return u.ID
}

// GetName returns the user name.
func (u User) GetName() string {
	return u.Name
}

func (u User) IsAdmin() bool {
	return u.Admin
}

func (u User) GetResources() []string {
	return u.Resources
}
