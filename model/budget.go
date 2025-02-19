package model

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BudgetRepository interface {
	Create(ctx context.Context, budget *Budget) error
	FindAll(ctx context.Context, query BudgetQueryInput) ([]Budget, int64, error)
	FindByID(ctx context.Context, id int64) (Budget, error)
	Update(ctx context.Context, id int64, budget Budget) error
	Delete(ctx context.Context, id int64) error
	FindOverviews(ctx context.Context, userID string) ([]BudgetOverview, error)
}

type Budget struct {
	ID               int64            `json:"id"`
	UserID           string           `json:"user_id"`
	Name             string           `json:"name"`
	Amount           decimal.Decimal  `json:"amount"`
	StartDate        string           `json:"start_date"`
	EndDate          string           `json:"end_date"`
	AutoRenew        bool             `json:"auto_renew"`
	RenewalPeriod    string           `json:"renewal_period"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `json:"deleted_at"`
	BudgetCategories []BudgetCategory `json:"budget_categories"`
}

type BudgetInput struct {
	UserID           string           `json:"user_id"`
	Name             string           `json:"name"`
	Amount           decimal.Decimal  `json:"amount"`
	StartDate        string           `json:"start_date"`
	EndDate          string           `json:"end_date"`
	AutoRenew        bool             `json:"auto_renew"`
	RenewalPeriod    string           `json:"renewal_period"`
	BudgetCategories []BudgetCategory `json:"budget_categories"`
}

type BudgetQueryInput struct {
	Keyword string `query:"keyword"`
	UserID  string `query:"user_id"`
	PaginatedRequest
}

type BudgetOverview struct {
	Budget
	TotalAmountTransaction decimal.Decimal `json:"total_amount_transaction"`
	Leftout                decimal.Decimal `json:"leftout"`
	Progress               decimal.Decimal `json:"progress"`
}

type BudgetCategory struct {
	ID        int64     `json:"id"`
	BudgetID  int64     `json:"budget_id"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt
}
