package utils

import (
	"testing"
)

func TestHashSourcePath_Uniqueness(t *testing.T) {
	testCases := []struct {
		name     string
		path1    string
		path2    string
		shouldEq bool
	}{
		{
			name:     "same path should produce same hash",
			path1:    "/home/user/.gitconfig",
			path2:    "/home/user/.gitconfig",
			shouldEq: true,
		},
		{
			name:     "different paths should produce different hashes",
			path1:    "/home/user/.gitconfig",
			path2:    "/home/user/.config/.gitconfig",
			shouldEq: false,
		},
		{
			name:     "nested paths should produce different hashes",
			path1:    "/home/user/.config/.gitconfig",
			path2:    "/home/user/.config/.gitignore",
			shouldEq: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash1 := HashSourcePath(tc.path1)
			hash2 := HashSourcePath(tc.path2)

			if tc.shouldEq && hash1 != hash2 {
				t.Errorf("Expected same hash for %s and %s, got %s and %s", tc.path1, tc.path2, hash1, hash2)
			}

			if !tc.shouldEq && hash1 == hash2 {
				t.Errorf("Expected different hashes for %s and %s, but both produced %s", tc.path1, tc.path2, hash1)
			}

			if len(hash1) != 16 {
				t.Errorf("Expected hash length 16, got %d for path %s", len(hash1), tc.path1)
			}

			if len(hash2) != 16 {
				t.Errorf("Expected hash length 16, got %d for path %s", len(hash2), tc.path2)
			}
		})
	}
}
