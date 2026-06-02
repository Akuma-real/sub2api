package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	refreshTokenKeyPrefix    = "refresh_token:"
	usedRefreshTokenPrefix   = "refresh_token_used:"
	userRefreshTokensPrefix  = "user_refresh_tokens:"
	tokenFamilyPrefix        = "token_family:"
	tokenFamilyRevokedPrefix = "token_family_revoked:"
	revokedTokenFamilyTTL    = 90 * 24 * time.Hour
)

// refreshTokenKey generates the Redis key for a refresh token.
func refreshTokenKey(tokenHash string) string {
	return refreshTokenKeyPrefix + tokenHash
}

// usedRefreshTokenKey generates the Redis key for a consumed refresh token.
func usedRefreshTokenKey(tokenHash string) string {
	return usedRefreshTokenPrefix + tokenHash
}

// userRefreshTokensKey generates the Redis key for user's token set.
func userRefreshTokensKey(userID int64) string {
	return fmt.Sprintf("%s%d", userRefreshTokensPrefix, userID)
}

// tokenFamilyKey generates the Redis key for token family set.
func tokenFamilyKey(familyID string) string {
	return tokenFamilyPrefix + familyID
}

// tokenFamilyRevokedKey generates the Redis key for a revoked token family marker.
func tokenFamilyRevokedKey(familyID string) string {
	return tokenFamilyRevokedPrefix + familyID
}

type refreshTokenCache struct {
	rdb *redis.Client
}

// NewRefreshTokenCache creates a new RefreshTokenCache implementation.
func NewRefreshTokenCache(rdb *redis.Client) service.RefreshTokenCache {
	return &refreshTokenCache{rdb: rdb}
}

func (c *refreshTokenCache) StoreRefreshToken(ctx context.Context, tokenHash string, data *service.RefreshTokenData, ttl time.Duration) error {
	key := refreshTokenKey(tokenHash)
	revokedKey := tokenFamilyRevokedKey(data.FamilyID)
	val, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal refresh token data: %w", err)
	}
	script := redis.NewScript(`
if redis.call("EXISTS", KEYS[2]) == 1 then
	return 0
end
redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
return 1
`)
	stored, err := script.Run(ctx, c.rdb, []string{key, revokedKey}, val, ttl.Milliseconds()).Int()
	if err != nil {
		return err
	}
	if stored == 0 {
		return service.ErrRefreshTokenReused
	}
	return nil
}

func (c *refreshTokenCache) StoreUsedRefreshToken(ctx context.Context, tokenHash string, data *service.RefreshTokenData, ttl time.Duration) error {
	key := usedRefreshTokenKey(tokenHash)
	val, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal refresh token data: %w", err)
	}
	stored, err := c.rdb.SetNX(ctx, key, val, ttl).Result()
	if err != nil {
		return err
	}
	if !stored {
		return service.ErrRefreshTokenReused
	}
	return nil
}

func (c *refreshTokenCache) GetRefreshToken(ctx context.Context, tokenHash string) (*service.RefreshTokenData, error) {
	key := refreshTokenKey(tokenHash)
	return c.getTokenData(ctx, key)
}

func (c *refreshTokenCache) GetUsedRefreshToken(ctx context.Context, tokenHash string) (*service.RefreshTokenData, error) {
	key := usedRefreshTokenKey(tokenHash)
	return c.getTokenData(ctx, key)
}

func (c *refreshTokenCache) getTokenData(ctx context.Context, key string) (*service.RefreshTokenData, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, service.ErrRefreshTokenNotFound
		}
		return nil, err
	}
	var data service.RefreshTokenData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("unmarshal refresh token data: %w", err)
	}
	return &data, nil
}

func (c *refreshTokenCache) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	key := refreshTokenKey(tokenHash)
	return c.rdb.Del(ctx, key).Err()
}

func (c *refreshTokenCache) DeleteUserRefreshTokens(ctx context.Context, userID int64) error {
	// Get all token hashes for this user
	tokenHashes, err := c.GetUserTokenHashes(ctx, userID)
	if err != nil && err != redis.Nil {
		return fmt.Errorf("get user token hashes: %w", err)
	}

	if len(tokenHashes) == 0 {
		return nil
	}

	// Build keys to delete
	keys := make([]string, 0, len(tokenHashes)*2+1)
	for _, hash := range tokenHashes {
		keys = append(keys, refreshTokenKey(hash))
		keys = append(keys, usedRefreshTokenKey(hash))
	}
	keys = append(keys, userRefreshTokensKey(userID))

	// Delete all keys in a pipeline
	pipe := c.rdb.Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (c *refreshTokenCache) DeleteTokenFamily(ctx context.Context, familyID string) error {
	// Get all token hashes in this family
	tokenHashes, err := c.GetFamilyTokenHashes(ctx, familyID)
	if err != nil && err != redis.Nil {
		return fmt.Errorf("get family token hashes: %w", err)
	}

	markerTTL, err := c.tokenFamilyRevocationTTL(ctx, tokenHashes)
	if err != nil {
		return fmt.Errorf("get family token revocation ttl: %w", err)
	}
	if markerTTL < revokedTokenFamilyTTL {
		markerTTL = revokedTokenFamilyTTL
	}

	if len(tokenHashes) == 0 {
		return c.rdb.Set(ctx, tokenFamilyRevokedKey(familyID), "1", markerTTL).Err()
	}

	// Build keys to delete
	keys := make([]string, 0, len(tokenHashes)*2+1)
	for _, hash := range tokenHashes {
		keys = append(keys, refreshTokenKey(hash))
		keys = append(keys, usedRefreshTokenKey(hash))
	}
	keys = append(keys, tokenFamilyKey(familyID))

	// Delete all keys in a pipeline
	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, tokenFamilyRevokedKey(familyID), "1", markerTTL)
	for _, key := range keys {
		pipe.Del(ctx, key)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (c *refreshTokenCache) tokenFamilyRevocationTTL(ctx context.Context, tokenHashes []string) (time.Duration, error) {
	now := time.Now()
	var ttl time.Duration
	for _, hash := range tokenHashes {
		data, err := c.GetRefreshToken(ctx, hash)
		if errors.Is(err, service.ErrRefreshTokenNotFound) {
			data, err = c.GetUsedRefreshToken(ctx, hash)
		}
		if errors.Is(err, service.ErrRefreshTokenNotFound) {
			continue
		}
		if err != nil {
			return 0, err
		}
		if remaining := data.ExpiresAt.Sub(now); remaining > ttl {
			ttl = remaining
		}
	}
	return ttl, nil
}

func (c *refreshTokenCache) IsTokenFamilyRevoked(ctx context.Context, familyID string) (bool, error) {
	count, err := c.rdb.Exists(ctx, tokenFamilyRevokedKey(familyID)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *refreshTokenCache) AddToUserTokenSet(ctx context.Context, userID int64, tokenHash string, ttl time.Duration) error {
	key := userRefreshTokensKey(userID)
	pipe := c.rdb.Pipeline()
	pipe.SAdd(ctx, key, tokenHash)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *refreshTokenCache) AddToFamilyTokenSet(ctx context.Context, familyID string, tokenHash string, ttl time.Duration) error {
	key := tokenFamilyKey(familyID)
	revokedKey := tokenFamilyRevokedKey(familyID)
	script := redis.NewScript(`
if redis.call("EXISTS", KEYS[2]) == 1 then
	return 0
end
redis.call("SADD", KEYS[1], ARGV[1])
redis.call("EXPIRE", KEYS[1], ARGV[2])
return 1
`)
	added, err := script.Run(ctx, c.rdb, []string{key, revokedKey}, tokenHash, int64(ttl.Seconds())).Int()
	if err != nil {
		return err
	}
	if added == 0 {
		return service.ErrRefreshTokenReused
	}
	return nil
}

func (c *refreshTokenCache) GetUserTokenHashes(ctx context.Context, userID int64) ([]string, error) {
	key := userRefreshTokensKey(userID)
	return c.rdb.SMembers(ctx, key).Result()
}

func (c *refreshTokenCache) GetFamilyTokenHashes(ctx context.Context, familyID string) ([]string, error) {
	key := tokenFamilyKey(familyID)
	return c.rdb.SMembers(ctx, key).Result()
}

func (c *refreshTokenCache) IsTokenInFamily(ctx context.Context, familyID string, tokenHash string) (bool, error) {
	key := tokenFamilyKey(familyID)
	return c.rdb.SIsMember(ctx, key, tokenHash).Result()
}
