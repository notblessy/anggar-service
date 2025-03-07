package model

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ScopeRepository interface {
	Create(ctx context.Context, scope ScopeInput) (Scope, error)
	FindAll(ctx context.Context, query ScopeQueryInput) ([]Scope, int64, error)
	FindByID(ctx context.Context, id int64) (Scope, error)
	Update(ctx context.Context, id int64, scope Scope) error
	Delete(ctx context.Context, id int64) error
	FindOverviews(ctx context.Context, userID string) ([]ScopeOverview, error)
}

type Scope struct {
	ID              int64           `json:"id"`
	UserID          string          `json:"user_id"`
	Name            string          `json:"name"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `json:"deleted_at"`
	ScopeCategories []ScopeCategory `json:"scope_categories"`
}

type ScopeInput struct {
	UserID        string          `json:"user_id"`
	Name          string          `json:"name"`
	Amount        decimal.Decimal `json:"amount"`
	StartDate     string          `json:"start_date"`
	EndDate       string          `json:"end_date"`
	AutoRenew     bool            `json:"auto_renew"`
	RenewalPeriod string          `json:"renewal_period"`
	CategoryIDs   []string        `json:"scope_categories"`
}

func (bi *ScopeInput) ToScope() Scope {
	return Scope{
		UserID: bi.UserID,
		Name:   bi.Name,
	}
}

type ScopeQueryInput struct {
	Keyword string `query:"keyword"`
	UserID  string `query:"user_id"`
	PaginatedRequest
}

type ScopeOverview struct {
	Scope
	TotalAmountTransaction decimal.Decimal `json:"total_amount_transaction"`
	Leftout                decimal.Decimal `json:"leftout"`
	Progress               decimal.Decimal `json:"progress"`
}

type ScopeCategory struct {
	ID        int64     `json:"id"`
	ScopeID   int64     `json:"scope_id"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt
}
