package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr, password string, db int) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStore{client: client}, nil
}

func (s *RedisStore) Client() *redis.Client {
	return s.client
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

// Auth code operations
func (s *RedisStore) SetAuthCode(ctx context.Context, phone, code string, expiration time.Duration) error {
	return s.client.Set(ctx, "auth_code:"+phone, code, expiration).Err()
}

func (s *RedisStore) GetAuthCode(ctx context.Context, phone string) (string, error) {
	return s.client.Get(ctx, "auth_code:"+phone).Result()
}

func (s *RedisStore) DeleteAuthCode(ctx context.Context, phone string) error {
	return s.client.Del(ctx, "auth_code:"+phone).Err()
}

// Session operations
func (s *RedisStore) SetSession(ctx context.Context, authKeyID string, userID int64, expiration time.Duration) error {
	return s.client.Set(ctx, "session:"+authKeyID, userID, expiration).Err()
}

func (s *RedisStore) GetSession(ctx context.Context, authKeyID string) (int64, error) {
	return s.client.Get(ctx, "session:"+authKeyID).Int64()
}

func (s *RedisStore) DeleteSession(ctx context.Context, authKeyID string) error {
	return s.client.Del(ctx, "session:"+authKeyID).Err()
}

// Online status operations
func (s *RedisStore) SetUserOnline(ctx context.Context, userID int64, expiration time.Duration) error {
	return s.client.Set(ctx, fmt.Sprintf("online:%d", userID), time.Now().Unix(), expiration).Err()
}

func (s *RedisStore) IsUserOnline(ctx context.Context, userID int64) (bool, error) {
	exists, err := s.client.Exists(ctx, fmt.Sprintf("online:%d", userID)).Result()
	return exists > 0, err
}

func (s *RedisStore) GetOnlineUsers(ctx context.Context) ([]int64, error) {
	keys, err := s.client.Keys(ctx, "online:*").Result()
	if err != nil {
		return nil, err
	}

	var userIDs []int64
	for _, key := range keys {
		var userID int64
		if _, err := fmt.Sscanf(key, "online:%d", &userID); err == nil {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs, nil
}

// Updates state operations
func (s *RedisStore) SetPts(ctx context.Context, userID int64, pts int) error {
	return s.client.Set(ctx, fmt.Sprintf("pts:%d", userID), pts, 0).Err()
}

func (s *RedisStore) GetPts(ctx context.Context, userID int64) (int, error) {
	return s.client.Get(ctx, fmt.Sprintf("pts:%d", userID)).Int()
}

func (s *RedisStore) IncrPts(ctx context.Context, userID int64) (int64, error) {
	return s.client.Incr(ctx, fmt.Sprintf("pts:%d", userID)).Result()
}

// Temp auth key binding
func (s *RedisStore) BindTempAuthKey(ctx context.Context, tempAuthKeyID, permAuthKeyID string, expiration time.Duration) error {
	return s.client.Set(ctx, "temp_bind:"+tempAuthKeyID, permAuthKeyID, expiration).Err()
}

func (s *RedisStore) GetPermAuthKeyID(ctx context.Context, tempAuthKeyID string) (string, error) {
	return s.client.Get(ctx, "temp_bind:"+tempAuthKeyID).Result()
}
