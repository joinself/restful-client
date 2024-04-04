package acl

import (
	"testing"
)

func TestIsAPermittedResource(t *testing.T) {
	tests := []struct {
		name      string
		permitted []string
		current   string
		want      bool
	}{
		{
			name:      "Wildcard on different resources",
			permitted: []string{"ANY /apps/710f4305-67f7-4ab5-a057-deb0e1c1ad52/connections/*/messages"},
			current:   "POST /apps/710f4305-67f7-4ab5-a057-deb0e1c1ad52/connections/78948918892/messages",
			want:      true,
		},

		{
			name:      "Exact match",
			permitted: []string{"GET /apps/resource1"},
			current:   "GET /apps/resource1",
			want:      true,
		},
		{
			name:      "Wildcard match",
			permitted: []string{"GET /apps/resource*"},
			current:   "GET /apps/resource2",
			want:      true,
		},
		{
			name:      "ANY match",
			permitted: []string{"ANY /apps/resource"},
			current:   "POST /apps/resource",
			want:      true,
		},
		{
			name:      "ANY wildcard match",
			permitted: []string{"ANY /apps/resource*"},
			current:   "POST /apps/resource2",
			want:      true,
		},
		{
			name:      "No match",
			permitted: []string{"GET /apps/resource1"},
			current:   "GET /apps/resource2",
			want:      false,
		},
		{
			name:      "ANY on different resources",
			permitted: []string{"ANY /apps/resource1"},
			current:   "POST /apps/resource2",
			want:      false,
		},
		{
			name:      "ANY on different method",
			permitted: []string{"POST /apps/resource1"},
			current:   "GET /apps/resource1",
			want:      false,
		},
		{
			name:      "Wildcard on different resources",
			permitted: []string{"GET /apps/resource1*"},
			current:   "GET /apps/resource2",
			want:      false,
		},
		{
			name:      "Exact match with trailing slashes",
			permitted: []string{"GET /apps/resource1/"},
			current:   "GET /apps/resource1/",
			want:      true,
		},
		{
			name:      "Mismatching trailing slashes",
			permitted: []string{"GET /apps/resource1"},
			current:   "GET /apps/resource1/",
			want:      true,
		},
		{
			name:      "Mismatching trailing slashes other way",
			permitted: []string{"GET /apps/resource1/"},
			current:   "GET /apps/resource1",
			want:      true,
		},
		{
			name:      "Mismatching leading slashes",
			permitted: []string{"GET /apps/resource1"},
			current:   "GET apps/resource1/",
			want:      false,
		},
		{
			name:      "ANY with invalid current path",
			permitted: []string{"ANY /apps/resource1/"},
			current:   "/apps/resource1/",
			want:      false,
		},
		{
			name:      "Wildcard with invalid template",
			permitted: []string{"/apps/resource*"},
			current:   "GET /apps/resource1/",
			want:      false,
		},
		{
			name:      "Medial-Wildcard template",
			permitted: []string{"GET /apps/resource/*/connections/A"},
			current:   "GET /apps/resource/resource1/connections/A",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAPermittedResource(tt.permitted, tt.current); got != tt.want {
				t.Errorf("isAPermittedResource() = %v, want %v", got, tt.want)
			}
		})
	}

}
