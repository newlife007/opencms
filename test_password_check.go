package main

import (
	"fmt"
	"github.com/openwan/media-asset-management/pkg/crypto"
)

func main() {
	// Test password from database
	hash := "$2a$10$diAaRVeFyZo582LYjzBt2.MX4TrchBmiJp6CU9gDRMBwMRzotKOPO"
	password := "admin123"
	
	fmt.Printf("Testing password: %s\n", password)
	fmt.Printf("Against hash: %s\n", hash)
	
	result := crypto.CheckPassword(password, hash)
	fmt.Printf("Result: %v\n", result)
	
	// Also test hashing
	newHash, err := crypto.HashPassword(password)
	if err != nil {
		fmt.Printf("Error hashing: %v\n", err)
	} else {
		fmt.Printf("New hash: %s\n", newHash)
		fmt.Printf("New hash check: %v\n", crypto.CheckPassword(password, newHash))
	}
}
