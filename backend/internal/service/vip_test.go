package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVIPServiceUpdateLevelClearsOriginalPrice(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewVIPService(client)

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
