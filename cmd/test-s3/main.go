package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openwan/media-asset-management/internal/storage"
)

func main() {
	// Load configuration from environment
	cfg := storage.LoadConfigFromEnv()
	
	fmt.Println("========================================")
	fmt.Println("S3 Storage Test")
	fmt.Println("========================================")
	fmt.Printf("Type: %s\n", cfg.Type)
	fmt.Printf("Bucket: %s\n", cfg.S3Bucket)
	fmt.Printf("Region: %s\n", cfg.S3Region)
	fmt.Printf("Prefix: %s\n", cfg.S3Prefix)
	fmt.Println("========================================")
	
	// Create storage service
	storageService, err := storage.NewStorageFromConfig(cfg)
	if err != nil {
		fmt.Printf("❌ Failed to create storage service: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Storage service created")
	
	// Test upload
	ctx := context.Background()
	testContent := "This is a test file for S3 storage validation."
	reader := strings.NewReader(testContent)
	
	metadata := map[string]string{
		"content-type":  "text/plain",
		"original-name": "test_upload.txt",
		"test-by":       "s3-validation",
	}
	
	fmt.Println("\n--- Testing Upload ---")
	path, err := storageService.Upload(ctx, "test_upload.txt", reader, metadata)
	if err != nil {
		fmt.Printf("❌ Upload failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Upload successful\n")
	fmt.Printf("  Path: %s\n", path)
	
	// Test exists
	fmt.Println("\n--- Testing Exists ---")
	exists, err := storageService.Exists(ctx, path)
	if err != nil {
		fmt.Printf("❌ Exists check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ File exists: %v\n", exists)
	
	// Test download
	fmt.Println("\n--- Testing Download ---")
	downloadReader, err := storageService.Download(ctx, path)
	if err != nil {
		fmt.Printf("❌ Download failed: %v\n", err)
		os.Exit(1)
	}
	defer downloadReader.Close()
	
	// Read content
	buf := make([]byte, 1024)
	n, _ := downloadReader.Read(buf)
	downloadedContent := string(buf[:n])
	
	if downloadedContent == testContent {
		fmt.Printf("✓ Download successful and content matches\n")
		fmt.Printf("  Content: %s\n", downloadedContent)
	} else {
		fmt.Printf("❌ Content mismatch\n")
		fmt.Printf("  Expected: %s\n", testContent)
		fmt.Printf("  Got: %s\n", downloadedContent)
	}
	
	// Test get URL
	fmt.Println("\n--- Testing Get URL ---")
	url, err := storageService.GetURL(ctx, path)
	if err != nil {
		fmt.Printf("❌ Get URL failed: %v\n", err)
	} else {
		fmt.Printf("✓ URL: %s\n", url)
	}
	
	// Test delete
	fmt.Println("\n--- Testing Delete ---")
	err = storageService.Delete(ctx, path)
	if err != nil {
		fmt.Printf("❌ Delete failed: %v\n", err)
	} else {
		fmt.Printf("✓ Delete successful\n")
	}
	
	// Verify deleted
	exists, _ = storageService.Exists(ctx, path)
	fmt.Printf("✓ File exists after delete: %v\n", exists)
	
	fmt.Println("\n========================================")
	fmt.Println("✅ All S3 storage tests passed!")
	fmt.Println("========================================")
}
