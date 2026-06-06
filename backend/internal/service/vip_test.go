package service

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

type vipAuthCacheInvalidatorStub struct {
	userIDs  []int64
	groupIDs []int64
	keys     []string
}

func (s *vipAuthCacheInvalidatorStub) InvalidateAuthCacheByKey(_ context.Context, key string) {
	s.keys = append(s.keys, key)
}

func (s *vipAuthCacheInvalidatorStub) InvalidateAuthCacheByUserID(_ context.Context, userID int64) {
	s.userIDs = append(s.userIDs, userID)
}

func (s *vipAuthCacheInvalidatorStub) InvalidateAuthCacheByGroupID(_ context.Context, groupID int64) {
	s.groupIDs = append(s.groupIDs, groupID)
}

func TestVIPServiceUpdateLevelClearsOriginalPrice(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewVIPService(client, nil)

	originalPrice := 199.0
	level, err := svc.CreateLevel(ctx, CreateVIPLevelRequest{
		Name:               "Gold",
		Price:              99,
		OriginalPrice:      &originalPrice,
		ValidityDays:       30,
		DiscountMultiplier: 0.8,
		ForSale:            true,
	})
	require.NoError(t, err)
	require.NotNil(t, level.OriginalPrice)

	level, err = svc.UpdateLevel(ctx, level.ID, UpdateVIPLevelRequest{
		ClearOriginalPrice: true,
	})
	require.NoError(t, err)
	require.Nil(t, level.OriginalPrice)
}

func TestVIPServiceAssignOrExtendVIPExtendsSameLevelAndInvalidatesCache(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	invalidator := &vipAuthCacheInvalidatorStub{}
	svc := NewVIPService(client, invalidator)

	user, err := client.User.Create().
		SetEmail("vip-assign@example.com").
		SetPasswordHash("hash").
		SetUsername("vip-assign-user").
		Save(ctx)
	require.NoError(t, err)

	level, err := svc.CreateLevel(ctx, CreateVIPLevelRequest{
		Name:               "1",
		Price:              19,
		ValidityDays:       30,
		DiscountMultiplier: 0.8,
		ForSale:            true,
	})
	require.NoError(t, err)

	first, extended, err := svc.AssignOrExtendVIP(ctx, AssignVIPInput{
		UserID:     user.ID,
		VIPLevelID: level.ID,
		Days:       10,
		Source:     "后台手动分配",
		Notes:      "initial grant",
	})
	require.NoError(t, err)
	require.False(t, extended)
	require.Equal(t, user.ID, first.UserID)
	require.Equal(t, level.ID, first.VipLevelID)
	require.NotNil(t, first.Edges.VipLevel)
	require.Equal(t, "后台手动分配: initial grant", derefVIPString(first.Notes))
	require.Equal(t, []int64{user.ID}, invalidator.userIDs)

	second, extended, err := svc.AssignOrExtendVIP(ctx, AssignVIPInput{
		UserID:     user.ID,
		VIPLevelID: level.ID,
		Days:       5,
		Source:     "兑换码兑换",
	})
	require.NoError(t, err)
	require.True(t, extended)
	require.Equal(t, first.ID, second.ID)
	require.WithinDuration(t, first.ExpiresAt.AddDate(0, 0, 5), second.ExpiresAt, time.Second)
	require.Contains(t, derefVIPString(second.Notes), "后台手动分配: initial grant")
	require.Contains(t, derefVIPString(second.Notes), "兑换码兑换")
	require.Equal(t, []int64{user.ID, user.ID}, invalidator.userIDs)
}

func TestVIPServiceAssignOrExtendVIPSkipsCacheInvalidationInsideTransaction(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	invalidator := &vipAuthCacheInvalidatorStub{}
	svc := NewVIPService(client, invalidator)

	user, err := client.User.Create().
		SetEmail("vip-tx@example.com").
		SetPasswordHash("hash").
		SetUsername("vip-tx-user").
		Save(ctx)
	require.NoError(t, err)

	level, err := svc.CreateLevel(ctx, CreateVIPLevelRequest{
		Name:               "Gold",
		Price:              19,
		ValidityDays:       30,
		DiscountMultiplier: 0.8,
		ForSale:            true,
	})
	require.NoError(t, err)

	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	txCtx := dbent.NewTxContext(ctx, tx)
	_, _, err = svc.AssignOrExtendVIP(txCtx, AssignVIPInput{
		UserID:     user.ID,
		VIPLevelID: level.ID,
		Days:       7,
		Source:     "兑换码兑换",
	})
	require.NoError(t, err)
	require.Empty(t, invalidator.userIDs)
	require.NoError(t, tx.Commit())
}
