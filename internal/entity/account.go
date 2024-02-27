package entity

import (
	"encoding/json"
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

type Scope struct {
	Resources []string `json:"resources"`
}

func (a *Account) GetResources() []string {
	var scope Scope
	err := json.Unmarshal([]byte(a.Resources), &scope)
	if err != nil {
		return strings.Split(a.Resources, ",")
	}
	return scope.Resources
}

func (a *Account) SetResources(resources []string) {
	s := Scope{
		Resources: resources,
	}
	b, err := json.Marshal(s)
	if err != nil {
		a.Resources = strings.Join(resources, ",")
	} else {
		a.Resources = string(b)
	}
}
