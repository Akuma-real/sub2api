package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenCache_UsedTokenMarkerAndFamilyRevocation(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		require.NoError(t, rdb.Close())
	})
	cache := NewRefreshTokenCache(rdb)

	data := &service.RefreshTokenData{
		UserID:       42,
		TokenVersion: 3,
		FamilyID:     "family-a",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	require.NoError(t, cache.StoreRefreshToken(ctx, "hash-a", data, time.Hour))
	require.NoError(t, cache.AddToFamilyTokenSet(ctx, data.FamilyID, "hash-a", time.Hour))
	require.NoError(t, cache.StoreUsedRefreshToken(ctx, "hash-a", data, time.Hour))

	err := cache.StoreUsedRefreshToken(ctx, "hash-a", data, time.Hour)
	require.ErrorIs(t, err, service.ErrRefreshTokenReused)

	require.NoError(t, cache.DeleteTokenFamily(ctx, data.FamilyID))

	revoked, err := cache.IsTokenFamilyRevoked(ctx, data.FamilyID)
	require.NoError(t, err)
	require.True(t, revoked)

	err = cache.StoreRefreshToken(ctx, "hash-b", data, time.Hour)
	require.True(t, errors.Is(err, service.ErrRefreshTokenReused), "revoked family must reject new tokens")

	err = cache.AddToFamilyTokenSet(ctx, data.FamilyID, "hash-b", time.Hour)
	require.True(t, errors.Is(err, service.ErrRefreshTokenReused), "revoked family must reject family membership")
}
