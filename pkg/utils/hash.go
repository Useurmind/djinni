package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashSourcePath generates a truncated SHA256 hash of the source path
// for use as a unique filename in the temp mount
func HashSourcePath(sourcePath string) string {
	hash := sha256.Sum256([]byte(sourcePath))
	return hex.EncodeToString(hash[:])[:16]
}
