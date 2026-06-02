package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestIsReservedEmail_DingTalkDomain(t *testing.T) {
	require.True(t, isReservedEmail("dingtalk-123@dingtalk-connect.invalid"))
	require.True(t, isReservedEmail("DINGTALK-456@DINGTALK-CONNECT.INVALID")) // case-insensitive
	require.False(t, isReservedEmail("real@dingtalk.com"))
}

func TestAuthServiceRefreshTokenPair_ReusedRefreshTokenRevokesFamily(t *testing.T) {
	ctx := context.Background()
	user := &User{
		ID:           42,
		Email:        "user@example.com",
		PasswordHash: "stable-password-hash",
		Role:         RoleUser,
		Status:       StatusActive,
		TokenVersion: 3,
	}
	cache := newRefreshTokenReplayCacheStub()
	svc := &AuthService{
		userRepo:          &refreshTokenReplayUserRepoStub{user: user},
		refreshTokenCache: cache,
		cfg: &config.Config{JWT: config.JWTConfig{
			Secret:                 "test-refresh-token-secret",
			ExpireHour:             1,
			RefreshTokenExpireDays: 7,
		}},
	}

	firstPair, err := svc.GenerateTokenPair(ctx, user, "")
	require.NoError(t, err)
	require.NotEmpty(t, firstPair.RefreshToken)

	secondPair, err := svc.RefreshTokenPair(ctx, firstPair.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, secondPair.RefreshToken)
	require.NotEqual(t, firstPair.RefreshToken, secondPair.RefreshToken)

	_, err = svc.RefreshTokenPair(ctx, firstPair.RefreshToken)
	require.ErrorIs(t, err, ErrRefreshTokenReused)

	_, err = svc.RefreshTokenPair(ctx, secondPair.RefreshToken)
	require.Error(t, err, "reusing an old refresh token must revoke the active descendant")
}

func TestAuthServiceRefreshTokenPair_RevokedFamilyRejectsStoredToken(t *testing.T) {
	ctx := context.Background()
	user := &User{
		ID:           42,
		Email:        "user@example.com",
		PasswordHash: "stable-password-hash",
		Role:         RoleUser,
		Status:       StatusActive,
		TokenVersion: 3,
	}
	cache := newRefreshTokenReplayCacheStub()
	svc := &AuthService{
		userRepo:          &refreshTokenReplayUserRepoStub{user: user},
		refreshTokenCache: cache,
		cfg: &config.Config{JWT: config.JWTConfig{
			Secret:                 "test-refresh-token-secret",
			ExpireHour:             1,
			RefreshTokenExpireDays: 7,
		}},
	}

	pair, err := svc.GenerateTokenPair(ctx, user, "")
	require.NoError(t, err)
	require.NotEmpty(t, pair.RefreshToken)

	tokenHash := hashToken(pair.RefreshToken)
	data, err := cache.GetRefreshToken(ctx, tokenHash)
	require.NoError(t, err)
	cache.revokedFamilies[data.FamilyID] = struct{}{}

	_, err = svc.RefreshTokenPair(ctx, pair.RefreshToken)
	require.ErrorIs(t, err, ErrRefreshTokenReused)

	_, err = cache.GetRefreshToken(ctx, tokenHash)
	require.ErrorIs(t, err, ErrRefreshTokenNotFound)
}

func TestAuthServiceRefreshTokenPair_UnknownRefreshTokenIsInvalid(t *testing.T) {
	ctx := context.Background()
	svc := &AuthService{
		refreshTokenCache: newRefreshTokenReplayCacheStub(),
		cfg: &config.Config{JWT: config.JWTConfig{
			Secret:                 "test-refresh-token-secret",
			ExpireHour:             1,
			RefreshTokenExpireDays: 7,
		}},
	}

	_, err := svc.RefreshTokenPair(ctx, "rt_unknown")
	require.ErrorIs(t, err, ErrRefreshTokenInvalid)
}

type refreshTokenReplayUserRepoStub struct {
	UserRepository
	user *User
}

func (s *refreshTokenReplayUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	if s.user == nil || s.user.ID != id {
		return nil, ErrUserNotFound
	}
	cloned := *s.user
	return &cloned, nil
}

type refreshTokenReplayCacheStub struct {
	tokens          map[string]*RefreshTokenData
	usedTokens      map[string]*RefreshTokenData
	userSets        map[int64]map[string]struct{}
	families        map[string]map[string]struct{}
	revokedFamilies map[string]struct{}
}

func newRefreshTokenReplayCacheStub() *refreshTokenReplayCacheStub {
	return &refreshTokenReplayCacheStub{
		tokens:          make(map[string]*RefreshTokenData),
		usedTokens:      make(map[string]*RefreshTokenData),
		userSets:        make(map[int64]map[string]struct{}),
		families:        make(map[string]map[string]struct{}),
		revokedFamilies: make(map[string]struct{}),
	}
}

func (s *refreshTokenReplayCacheStub) StoreRefreshToken(_ context.Context, tokenHash string, data *RefreshTokenData, _ time.Duration) error {
	if _, revoked := s.revokedFamilies[data.FamilyID]; revoked {
		return ErrRefreshTokenReused
	}
	cloned := *data
	s.tokens[tokenHash] = &cloned
	return nil
}

func (s *refreshTokenReplayCacheStub) GetRefreshToken(_ context.Context, tokenHash string) (*RefreshTokenData, error) {
	data, ok := s.tokens[tokenHash]
	if !ok {
		return nil, ErrRefreshTokenNotFound
	}
	cloned := *data
	return &cloned, nil
}

func (s *refreshTokenReplayCacheStub) DeleteRefreshToken(_ context.Context, tokenHash string) error {
	delete(s.tokens, tokenHash)
	return nil
}

func (s *refreshTokenReplayCacheStub) StoreUsedRefreshToken(_ context.Context, tokenHash string, data *RefreshTokenData, _ time.Duration) error {
	if _, exists := s.usedTokens[tokenHash]; exists {
		return ErrRefreshTokenReused
	}
	cloned := *data
	s.usedTokens[tokenHash] = &cloned
	return nil
}

func (s *refreshTokenReplayCacheStub) GetUsedRefreshToken(_ context.Context, tokenHash string) (*RefreshTokenData, error) {
	data, ok := s.usedTokens[tokenHash]
	if !ok {
		return nil, ErrRefreshTokenNotFound
	}
	cloned := *data
	return &cloned, nil
}

func (s *refreshTokenReplayCacheStub) DeleteUserRefreshTokens(_ context.Context, userID int64) error {
	for tokenHash := range s.userSets[userID] {
		s.deleteTokenFromSets(tokenHash)
	}
	delete(s.userSets, userID)
	return nil
}

func (s *refreshTokenReplayCacheStub) DeleteTokenFamily(_ context.Context, familyID string) error {
	s.revokedFamilies[familyID] = struct{}{}
	for tokenHash := range s.families[familyID] {
		s.deleteTokenFromSets(tokenHash)
	}
	delete(s.families, familyID)
	return nil
}

func (s *refreshTokenReplayCacheStub) IsTokenFamilyRevoked(_ context.Context, familyID string) (bool, error) {
	_, ok := s.revokedFamilies[familyID]
	return ok, nil
}

func (s *refreshTokenReplayCacheStub) AddToUserTokenSet(_ context.Context, userID int64, tokenHash string, _ time.Duration) error {
	if s.userSets[userID] == nil {
		s.userSets[userID] = make(map[string]struct{})
	}
	s.userSets[userID][tokenHash] = struct{}{}
	return nil
}

func (s *refreshTokenReplayCacheStub) AddToFamilyTokenSet(_ context.Context, familyID string, tokenHash string, _ time.Duration) error {
	if _, revoked := s.revokedFamilies[familyID]; revoked {
		delete(s.tokens, tokenHash)
		delete(s.usedTokens, tokenHash)
		return ErrRefreshTokenReused
	}
	if s.families[familyID] == nil {
		s.families[familyID] = make(map[string]struct{})
	}
	s.families[familyID][tokenHash] = struct{}{}
	return nil
}

func (s *refreshTokenReplayCacheStub) GetUserTokenHashes(_ context.Context, userID int64) ([]string, error) {
	tokenSet := s.userSets[userID]
	out := make([]string, 0, len(tokenSet))
	for tokenHash := range tokenSet {
		out = append(out, tokenHash)
	}
	return out, nil
}

func (s *refreshTokenReplayCacheStub) GetFamilyTokenHashes(_ context.Context, familyID string) ([]string, error) {
	tokenSet := s.families[familyID]
	out := make([]string, 0, len(tokenSet))
	for tokenHash := range tokenSet {
		out = append(out, tokenHash)
	}
	return out, nil
}

func (s *refreshTokenReplayCacheStub) IsTokenInFamily(_ context.Context, familyID string, tokenHash string) (bool, error) {
	_, ok := s.families[familyID][tokenHash]
	return ok, nil
}

func (s *refreshTokenReplayCacheStub) deleteTokenFromSets(tokenHash string) {
	delete(s.tokens, tokenHash)
	delete(s.usedTokens, tokenHash)
	for _, tokenSet := range s.userSets {
		delete(tokenSet, tokenHash)
	}
	for _, tokenSet := range s.families {
		delete(tokenSet, tokenHash)
	}
}
