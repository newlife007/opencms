package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a password with a hash
// Supports bcrypt (new format), MD5-crypt (legacy $1$), and plain MD5 (legacy format)
func CheckPassword(password, hash string) bool {
	// Check if it's a bcrypt hash (starts with $2a$, $2b$, or $2y$)
	if strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$") || strings.HasPrefix(hash, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		return err == nil
	}

	// Check if it's MD5-crypt format (starts with $1$) from legacy PHP system
	if strings.HasPrefix(hash, "$1$") {
		c := crypt.MD5.New()
		err := c.Verify(hash, []byte(password))
		return err == nil
	}

	// Check if it's plain MD5 (32 character hex string from legacy PHP system)
	if len(hash) == 32 {
		// Compute MD5 of the password
		h := md5.New()
		h.Write([]byte(password))
		computed := hex.EncodeToString(h.Sum(nil))
		return computed == hash
	}

	// Unknown hash format
	return false
}

// HashPasswordMD5 generates MD5 hash (for legacy compatibility)
func HashPasswordMD5(password string) string {
	h := md5.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}
