package entity

import "time"

const (
	APP_CREATED_STATUS  = "created"
	APP_ENABLED_STATUS  = "enabled"
	APP_DISABLED_STATUS = "disabled"
	APP_CRASHED_STATUS  = "crashed"
)

// App represents an app record.
type App struct {
	// AppID is the Self app self identifier.
	ID string `json:"id"`
	// DeviceSecret is the secret key for the device created at the developer portal.
	DeviceSecret string `json:"device_secret,omitempty"`
	// Name is the Self app name.
	Name string `json:"name"`
	// Env is self environment you want to point to, when empty, it will default to production.
	Env string `json:"env"`
	// Callback is the url that will be hit when a message is received.
	Callback string `json:"callback"`
	// CallbackSecret is the secret that will be used to sign your callbacks.
	CallbackSecret string    `json:"callback_secret,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
