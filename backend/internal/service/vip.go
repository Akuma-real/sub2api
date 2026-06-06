package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/ent/usagelog"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/ent/uservipmembership"
	"github.com/Wei-Shaw/sub2api/ent/viplevel"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	VIPMembershipStatusActive  = "active"
	VIPMembershipStatusExpired = "expired"
	VIPMembershipStatusRevoked = "revoked"
)

type VIPLevel struct {
	ID                 int64          `json:"id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	Price              float64        `json:"price"`
	OriginalPrice      *float64       `json:"original_price,omitempty"`
	ValidityDays       int            `json:"validity_days"`
	DiscountMultiplier float64        `json:"discount_multiplier"`
	Features           string         `json:"features"`
	Benefits           map[string]any `json:"benefits"`
	ForSale            bool           `json:"for_sale"`
	SortOrder          int            `json:"sort_order"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type UserVIPMembership struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	VIPLevelID    int64     `json:"vip_level_id"`
	StartsAt      time.Time `json:"starts_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	Status        string    `json:"status"`
	SourceOrderID *int64    `json:"source_order_id,omitempty"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Level         *VIPLevel `json:"level,omitempty"`
}

type VIPOverview struct {
	Current      *UserVIPMembership   `json:"current"`
	History      []*UserVIPMembership `json:"history,omitempty"`
	TotalSavings float64              `json:"total_savings_usd"`
}

type VIPUserSummary struct {
	UserID       int64              `json:"user_id"`
	Email        string             `json:"email"`
	Username     string             `json:"username"`
	Current      *UserVIPMembership `json:"current,omitempty"`
	TotalSavings float64            `json:"total_savings_usd"`
}

type CreateVIPLevelRequest struct {
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	Price              float64        `json:"price"`
	OriginalPrice      *float64       `json:"original_price"`
	ValidityDays       int            `json:"validity_days"`
	DiscountMultiplier float64        `json:"discount_multiplier"`
	Features           string         `json:"features"`
	Benefits           map[string]any `json:"benefits"`
	ForSale            bool           `json:"for_sale"`
	SortOrder          int            `json:"sort_order"`
}

type UpdateVIPLevelRequest struct {
	Name               *string         `json:"name"`
	Description        *string         `json:"description"`
	Price              *float64        `json:"price"`
	OriginalPrice      *float64        `json:"original_price"`
	ClearOriginalPrice bool            `json:"clear_original_price"`
	ValidityDays       *int            `json:"validity_days"`
	DiscountMultiplier *float64        `json:"discount_multiplier"`
	Features           *string         `json:"features"`
	Benefits           *map[string]any `json:"benefits"`
	ForSale            *bool           `json:"for_sale"`
	SortOrder          *int            `json:"sort_order"`
}

type VIPService struct {
	entClient *dbent.Client
}

func NewVIPService(entClient *dbent.Client) *VIPService {
	return &VIPService{entClient: entClient}
}

func (s *VIPService) ListLevels(ctx context.Context, forSaleOnly bool) ([]*dbent.VIPLevel, error) {
	q := s.entClient.VIPLevel.Query().Order(viplevel.BySortOrder(), viplevel.ByID())
	if forSaleOnly {
		q = q.Where(viplevel.ForSaleEQ(true))
	}
	return q.All(ctx)
}

func (s *VIPService) GetLevel(ctx context.Context, id int64) (*dbent.VIPLevel, error) {
	level, err := s.entClient.VIPLevel.Get(ctx, id)
	if err != nil {
		return nil, infraerrors.NotFound("VIP_LEVEL_NOT_FOUND", "vip level not found")
	}
	return level, nil
}

func (s *VIPService) CreateLevel(ctx context.Context, req CreateVIPLevelRequest) (*dbent.VIPLevel, error) {
	if err := validateVIPLevel(req.Name, req.Price, req.ValidityDays, req.DiscountMultiplier, req.OriginalPrice); err != nil {
		return nil, err
	}
	benefits := req.Benefits
	if benefits == nil {
		benefits = map[string]any{}
	}
	b := s.entClient.VIPLevel.Create().
		SetName(strings.TrimSpace(req.Name)).
		SetDescription(req.Description).
		SetPrice(req.Price).
		SetValidityDays(req.ValidityDays).
		SetDiscountMultiplier(req.DiscountMultiplier).
		SetFeatures(req.Features).
		SetBenefits(benefits).
		SetForSale(req.ForSale).
		SetSortOrder(req.SortOrder)
	if req.OriginalPrice != nil {
		b.SetOriginalPrice(*req.OriginalPrice)
	}
	return b.Save(ctx)
}

func (s *VIPService) UpdateLevel(ctx context.Context, id int64, req UpdateVIPLevelRequest) (*dbent.VIPLevel, error) {
	if err := validateVIPLevelPatch(req); err != nil {
		return nil, err
	}
	u := s.entClient.VIPLevel.UpdateOneID(id)
	if req.Name != nil {
		u.SetName(strings.TrimSpace(*req.Name))
	}
	if req.Description != nil {
		u.SetDescription(*req.Description)
	}
	if req.Price != nil {
		u.SetPrice(*req.Price)
	}
	if req.ClearOriginalPrice {
		u.ClearOriginalPrice()
	} else if req.OriginalPrice != nil {
		u.SetOriginalPrice(*req.OriginalPrice)
	}
	if req.ValidityDays != nil {
		u.SetValidityDays(*req.ValidityDays)
	}
	if req.DiscountMultiplier != nil {
		u.SetDiscountMultiplier(*req.DiscountMultiplier)
	}
	if req.Features != nil {
		u.SetFeatures(*req.Features)
	}
	if req.Benefits != nil {
		u.SetBenefits(*req.Benefits)
	}
	if req.ForSale != nil {
		u.SetForSale(*req.ForSale)
	}
	if req.SortOrder != nil {
		u.SetSortOrder(*req.SortOrder)
	}
	level, err := u.Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, infraerrors.NotFound("VIP_LEVEL_NOT_FOUND", "vip level not found")
		}
		return nil, err
	}
	return level, nil
}

func (s *VIPService) DeleteLevel(ctx context.Context, id int64) error {
	pending, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.VipLevelIDEQ(id),
			paymentorder.StatusIn(pendingOrderStatuses...),
		).Count(ctx)
	if err != nil {
		return err
	}
	if pending > 0 {
		return infraerrors.Conflict("PENDING_ORDERS", fmt.Sprintf("this VIP level has %d in-progress orders", pending))
	}
	if err := s.entClient.VIPLevel.DeleteOneID(id).Exec(ctx); err != nil {
		if dbent.IsNotFound(err) {
			return infraerrors.NotFound("VIP_LEVEL_NOT_FOUND", "vip level not found")
		}
		return err
	}
	return nil
}

func (s *VIPService) GetCurrentMembership(ctx context.Context, userID int64) (*dbent.UserVIPMembership, error) {
	now := time.Now()
	m, err := s.entClient.UserVIPMembership.Query().
		Where(
			uservipmembership.UserIDEQ(userID),
			uservipmembership.StatusEQ(VIPMembershipStatusActive),
			uservipmembership.ExpiresAtGT(now),
		).
		WithVipLevel().
		Order(uservipmembership.ByVipLevelField(viplevel.FieldDiscountMultiplier), uservipmembership.ByExpiresAt(entsql.OrderDesc())).
		First(ctx)
	if dbent.IsNotFound(err) {
		return nil, nil
	}
	return m, err
}

func (s *VIPService) GetOverview(ctx context.Context, userID int64) (*VIPOverview, error) {
	current, err := s.GetCurrentMembership(ctx, userID)
	if err != nil {
		return nil, err
	}
	history, err := s.entClient.UserVIPMembership.Query().
		Where(uservipmembership.UserIDEQ(userID)).
		WithVipLevel().
		Order(uservipmembership.ByCreatedAt(entsql.OrderDesc())).
		Limit(20).
		All(ctx)
	if err != nil {
		return nil, err
	}
	total, err := s.GetTotalSavings(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &VIPOverview{Current: toServiceVIPMembership(current), History: toServiceVIPMemberships(history), TotalSavings: total}, nil
}

func (s *VIPService) GetTotalSavings(ctx context.Context, userID int64) (float64, error) {
	sum, err := s.entClient.UsageLog.Query().
		Where(usagelog.UserIDEQ(userID)).
		Aggregate(dbent.Sum(usagelog.FieldVipSavingsUsd)).
		Float64(ctx)
	if dbent.IsNotFound(err) {
		return 0, nil
	}
	return sum, err
}

func (s *VIPService) ListUserSummaries(ctx context.Context, page, pageSize int) ([]VIPUserSummary, int, error) {
	ps, pg := applyPagination(pageSize, page)
	users, err := s.entClient.User.Query().
		Order(user.ByID()).
		Limit(ps).
		Offset((pg - 1) * ps).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.entClient.User.Query().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]VIPUserSummary, 0, len(users))
	for _, u := range users {
		current, _ := s.GetCurrentMembership(ctx, u.ID)
		savings, _ := s.GetTotalSavings(ctx, u.ID)
		out = append(out, VIPUserSummary{
			UserID:       u.ID,
			Email:        u.Email,
			Username:     u.Username,
			Current:      toServiceVIPMembership(current),
			TotalSavings: savings,
		})
	}
	return out, total, nil
}

func validateVIPLevel(name string, price float64, days int, multiplier float64, originalPrice *float64) error {
	if strings.TrimSpace(name) == "" {
		return infraerrors.BadRequest("VIP_LEVEL_NAME_REQUIRED", "vip level name is required")
	}
	if price <= 0 {
		return infraerrors.BadRequest("VIP_LEVEL_PRICE_INVALID", "vip price must be > 0")
	}
	if days <= 0 {
		return infraerrors.BadRequest("VIP_LEVEL_VALIDITY_INVALID", "vip validity days must be > 0")
	}
	if multiplier <= 0 || multiplier > 1 {
		return infraerrors.BadRequest("VIP_LEVEL_DISCOUNT_INVALID", "vip discount multiplier must be in (0, 1]")
	}
	if originalPrice != nil && *originalPrice < 0 {
		return infraerrors.BadRequest("VIP_LEVEL_ORIGINAL_PRICE_INVALID", "original price must be >= 0")
	}
	return nil
}

func validateVIPLevelPatch(req UpdateVIPLevelRequest) error {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return infraerrors.BadRequest("VIP_LEVEL_NAME_REQUIRED", "vip level name is required")
	}
	if req.Price != nil && *req.Price <= 0 {
		return infraerrors.BadRequest("VIP_LEVEL_PRICE_INVALID", "vip price must be > 0")
	}
	if req.ValidityDays != nil && *req.ValidityDays <= 0 {
		return infraerrors.BadRequest("VIP_LEVEL_VALIDITY_INVALID", "vip validity days must be > 0")
	}
	if req.DiscountMultiplier != nil && (*req.DiscountMultiplier <= 0 || *req.DiscountMultiplier > 1) {
		return infraerrors.BadRequest("VIP_LEVEL_DISCOUNT_INVALID", "vip discount multiplier must be in (0, 1]")
	}
	if !req.ClearOriginalPrice && req.OriginalPrice != nil && *req.OriginalPrice < 0 {
		return infraerrors.BadRequest("VIP_LEVEL_ORIGINAL_PRICE_INVALID", "original price must be >= 0")
	}
	return nil
}

func toServiceVIPLevel(level *dbent.VIPLevel) *VIPLevel {
	if level == nil {
		return nil
	}
	return &VIPLevel{
		ID:                 level.ID,
		Name:               level.Name,
		Description:        level.Description,
		Price:              level.Price,
		OriginalPrice:      level.OriginalPrice,
		ValidityDays:       level.ValidityDays,
		DiscountMultiplier: level.DiscountMultiplier,
		Features:           level.Features,
		Benefits:           level.Benefits,
		ForSale:            level.ForSale,
		SortOrder:          level.SortOrder,
		CreatedAt:          level.CreatedAt,
		UpdatedAt:          level.UpdatedAt,
	}
}

func toServiceVIPMembership(m *dbent.UserVIPMembership) *UserVIPMembership {
	if m == nil {
		return nil
	}
	notes := ""
	if m.Notes != nil {
		notes = *m.Notes
	}
	return &UserVIPMembership{
		ID:            m.ID,
		UserID:        m.UserID,
		VIPLevelID:    m.VipLevelID,
		StartsAt:      m.StartsAt,
		ExpiresAt:     m.ExpiresAt,
		Status:        m.Status,
		SourceOrderID: m.SourceOrderID,
		Notes:         notes,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		Level:         toServiceVIPLevel(m.Edges.VipLevel),
	}
}

func toServiceVIPMemberships(rows []*dbent.UserVIPMembership) []*UserVIPMembership {
	out := make([]*UserVIPMembership, 0, len(rows))
	for _, row := range rows {
		out = append(out, toServiceVIPMembership(row))
	}
	return out
}
