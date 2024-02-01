package entity

import (
	"strings"
	"time"
)

type Account struct {
	ID                     int       `json:"id"`
	UserName               string    `json:"user_name"`
	HashedPassword         string    `json:"hashed_password"`
	Password               string    `json:"password" db:"-"`
	Salt                   string    `json:"salt"`
	Resources              string    `json:"resources"`
	RequiresPasswordChange int       `json:"requires_password_change"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (a *Account) GetResources() []string {
	return strings.Split(a.Resources, ",")
}

func (a *Account) SetResources(resources []string) {
	a.Resources = strings.Join(resources, ",")
}
