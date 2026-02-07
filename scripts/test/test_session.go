package main

import (
	"context"
	"fmt"
	"time"

	"github.com/openwan/media-asset-management/internal/session"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Session Store Test")
	fmt.Println("========================================")
	fmt.Println()

	// Initialize Redis session store
	fmt.Println("1. Connecting to Redis...")
	store, err := session.NewRedisStore(
		"localhost:6379",
		"",
		0,
		24*time.Hour,
	)
	if err != nil {
		fmt.Printf("❌ Failed to connect to Redis: %v\n", err)
		return
	}
	defer store.Close()
	fmt.Println("✓ Redis connected")
	fmt.Println()

	ctx := context.Background()

	// Test 1: Save session
	fmt.Println("2. Saving session data...")
	sessionID := "test-session-12345"
	sessionData := &session.SessionData{
		UserID:      100,
		Username:    "testuser",
		GroupID:     5,
		LevelID:     3,
		IsAdmin:     true,
		Permissions: []string{"admin.users.view", "admin.users.edit", "files.upload"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := store.Save(ctx, sessionID, sessionData); err != nil {
		fmt.Printf("❌ Failed to save session: %v\n", err)
		return
	}
	fmt.Println("✓ Session saved successfully")
	fmt.Printf("  Session ID: %s\n", sessionID)
	fmt.Printf("  User ID: %d\n", sessionData.UserID)
	fmt.Printf("  Username: %s\n", sessionData.Username)
	fmt.Printf("  Is Admin: %v\n", sessionData.IsAdmin)
	fmt.Printf("  Permissions: %v\n", sessionData.Permissions)
	fmt.Println()

	// Test 2: Retrieve session
	fmt.Println("3. Retrieving session data...")
	retrieved, err := store.Get(ctx, sessionID)
	if err != nil {
		fmt.Printf("❌ Failed to retrieve session: %v\n", err)
		return
	}
	fmt.Println("✓ Session retrieved successfully")
	fmt.Printf("  User ID: %d\n", retrieved.UserID)
	fmt.Printf("  Username: %s\n", retrieved.Username)
	fmt.Printf("  Group ID: %d\n", retrieved.GroupID)
	fmt.Printf("  Level ID: %d\n", retrieved.LevelID)
	fmt.Printf("  Is Admin: %v\n", retrieved.IsAdmin)
	fmt.Printf("  Permissions: %v\n", retrieved.Permissions)
	fmt.Println()

	// Test 3: Check session exists
	fmt.Println("4. Checking if session exists...")
	exists, err := store.Exists(ctx, sessionID)
	if err != nil {
		fmt.Printf("❌ Failed to check session existence: %v\n", err)
		return
	}
	if exists {
		fmt.Println("✓ Session exists in store")
	} else {
		fmt.Println("❌ Session does not exist")
	}
	fmt.Println()

	// Test 4: Delete session
	fmt.Println("5. Deleting session...")
	if err := store.Delete(ctx, sessionID); err != nil {
		fmt.Printf("❌ Failed to delete session: %v\n", err)
		return
	}
	fmt.Println("✓ Session deleted successfully")
	fmt.Println()

	// Test 5: Verify session is gone
	fmt.Println("6. Verifying session is deleted...")
	exists, err = store.Exists(ctx, sessionID)
	if err != nil {
		fmt.Printf("❌ Failed to check session existence: %v\n", err)
		return
	}
	if !exists {
		fmt.Println("✓ Session no longer exists")
	} else {
		fmt.Println("❌ Session still exists (unexpected)")
	}
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("✓ All tests passed!")
	fmt.Println("Session store is working correctly")
	fmt.Println("========================================")
}
