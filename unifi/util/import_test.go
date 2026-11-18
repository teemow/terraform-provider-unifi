package util

import "testing"

func TestIsObjectID(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		// Network 10.x and older: MongoDB ObjectID.
		"0123456789abcdef01234567": true,
		// Network 11: UUID.
		"01234567-89ab-cdef-0123-456789abcdef": true,
		// Everything else is a name, which the import functions look up.
		"Corporate LAN":                         false,
		"default":                               false,
		"":                                      false,
		"0123456789abcdef0123456":               false,
		"0123456789ABCDEF01234567":              false,
		"01234567-89ab-cdef-0123-456789abcde":   false,
		"0123456789abcdef0123456789abcdef":      false,
		"x01234567-89ab-cdef-0123-456789abcdef": false,
	}

	for id, want := range tests {
		if got := IsObjectID(id); got != want {
			t.Errorf("IsObjectID(%q) = %v, want %v", id, got, want)
		}
	}
}
