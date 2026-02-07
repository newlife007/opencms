package transcoding

import (
	"fmt"
	"os"
	"syscall"
)

// FileLock represents a file-based lock
type FileLock struct {
	path string
	file *os.File
}

// NewFileLock creates a new file lock
func NewFileLock(path string) *FileLock {
	return &FileLock{path: path}
}

// Lock acquires the lock
func (l *FileLock) Lock() error {
	// Create lock file
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}
	
	// Acquire exclusive lock using flock
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	
	l.file = file
	return nil
}

// Unlock releases the lock
func (l *FileLock) Unlock() error {
	if l.file == nil {
		return nil
	}
	
	// Release lock
	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	
	// Close and remove lock file
	l.file.Close()
	os.Remove(l.path)
	
	return nil
}

// IsLocked checks if the lock file exists
func IsLocked(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
