package response

// Error used to respond to errored http requests.
type Error struct {
	Status  int    `json:"status"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Path    string `json:"path,omitempty"`
}
