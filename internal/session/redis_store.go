package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store is the interface for session storage
type Store interface {
	Save(ctx context.Context, sessionID string, data *SessionData) error
	Get(ctx context.Context, sessionID string) (*SessionData, error)
	Delete(ctx context.Context, sessionID string) error
	Exists(ctx context.Context, sessionID string) (bool, error)
	Close() error
}

// Session represents a user session
type Session struct {
	ID   string
	Data map[string]interface{}
}

// SessionData holds session information
type SessionData struct {
	UserID      int                    `json:"user_id"`
	Username    string                 `json:"username"`
	GroupID     int                    `json:"group_id"`
	LevelID     int                    `json:"level_id"`
	IsAdmin     bool                   `json:"is_admin"`
	Permissions []string               `json:"permissions"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RedisStore implements session storage using Redis
type RedisStore struct {
	client  *redis.Client
	ttl     time.Duration
	keyPrefix string
}

// NewRedisStore creates a new Redis session store
func NewRedisStore(addr, password string, db int, ttl time.Duration) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	
	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	return &RedisStore{
		client:    client,
		ttl:       ttl,
		keyPrefix: "session:",
	}, nil
}

// Save stores session data in Redis
func (s *RedisStore) Save(ctx context.Context, sessionID string, data *SessionData) error {
	data.UpdatedAt = time.Now()
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	
	key := s.keyPrefix + sessionID
	if err := s.client.Set(ctx, key, jsonData, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	
	return nil
}

// Get retrieves session data from Redis
func (s *RedisStore) Get(ctx context.Context, sessionID string) (*SessionData, error) {
	key := s.keyPrefix + sessionID
	
	jsonData, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	var data SessionData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}
	
	// Refresh TTL on access
	s.client.Expire(ctx, key, s.ttl)
	
	return &data, nil
}

// Delete removes a session from Redis
func (s *RedisStore) Delete(ctx context.Context, sessionID string) error {
	key := s.keyPrefix + sessionID
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// Exists checks if a session exists
func (s *RedisStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := s.keyPrefix + sessionID
	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Close closes the Redis connection
func (s *RedisStore) Close() error {
	return s.client.Close()
}
