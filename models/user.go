package models

type User struct {
	ID       uint64 `json:"id,omitempty" db:"id"`
	Name     string `json:"name,omitempty" db:"name"`
	LastName string `json:"last_name,omitempty" db:"last_name"`
	Email    string `json:"email,omitempty" db:"email"`
	Password []byte `json:"_" db:"password"`
}
