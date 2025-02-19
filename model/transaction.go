package model

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	TransactionTypeIncome  = "income"
	TransactionTypeExpense = "expense"

	CategoryOpname = "opname"
)

type TransactionRepository interface {
	Create(c context.Context, transaction *Transaction) error
	FindAll(c context.Context, query TransactionQueryInput) ([]Transaction, int64, error)
	FindByID(c context.Context, id int64) (Transaction, error)
	Update(c context.Context, id int64, transaction Transaction) error
	Delete(c context.Context, id int64) error
}

type Transaction struct {
	ID              int64           `json:"id"`
	UserID          string          `json:"user_id"`
	WalletID        int64           `json:"wallet_id"`
	BudgetID        int64           `json:"budget_id"`
	Category        string          `json:"category"`
	TransactionType string          `json:"transaction_type"`
	Description     string          `json:"description"`
	SpentAt         time.Time       `json:"spent_at"`
	Amount          decimal.Decimal `json:"amount"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `json:"deleted_at"`
}

type TransactionInput struct {
	UserID      string          `json:"user_id"`
	WalletID    int64           `json:"wallet_id"`
	BudgetID    int64           `json:"budget_id"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
	SpentAt     time.Time       `json:"spent_at"`
	Amount      decimal.Decimal `json:"amount"`
}

type TransactionQueryInput struct {
	Keyword string `query:"keyword"`
	PaginatedRequest
}
