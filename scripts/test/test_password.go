package main

import (
	"fmt"
	"github.com/openwan/media-asset-management/pkg/crypto"
)

func main() {
	hash := "$1$kI0.dK0.$mZfeLOhcTZ.xHq5uw8fk3."
	
	// Test common passwords
	passwords := []string{
		"admin123",
		"admin",
		"123456",
		"password",
		"openwan",
	}
	
	fmt.Println("Testing password hash:", hash)
	fmt.Println()
	
	for _, pwd := range passwords {
		result := crypto.CheckPassword(pwd, hash)
		if result {
			fmt.Printf("✓ Password MATCH: %s\n", pwd)
		} else {
			fmt.Printf("✗ Password FAILED: %s\n", pwd)
		}
	}
}
