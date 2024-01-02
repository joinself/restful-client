package entity

import "time"

// App represents an app record.
type App struct {
	// AppID is the Self app self identifier.
	ID string `json:"id"`
	// DeviceSecret is the secret key for the device created at the developer portal.
	DeviceSecret string `json:"device_secret"`
	// Name is the Self app name.
	Name string `json:"name"`
	// Env is self environment you want to point to, when empty, it will default to production.
	Env string `json:"env"`
	// Callback is the url that will be hit when a message is received.
	Callback  string    `json:"callback"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
