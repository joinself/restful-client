package entity

import (
	"time"
)

// Room represents a message group.
type Room struct {
	ID        int       `json:"id"`
	Appid     string    `json:"appid"`
	GID       string    `json:"gid"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	IconLink  string    `json:"icon_link"`
	IconMime  string    `json:"icon_mime"`
	IconKey   string    `json:"icon_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RoomConnection struct {
	ID           int       `json:"id"`
	RoomID       int       `json:"room_id"`
	ConnectionID int       `json:"connection_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
