package helper

import "strings"

// FlattenSelfID removes anything after the colon, removing any references
// to a device on the self identifier..
func FlattenSelfID(selfID string) string {
	parts := strings.Split(selfID, ":")
	if len(parts) > 0 {
		return parts[0]
	}

	return selfID
}
