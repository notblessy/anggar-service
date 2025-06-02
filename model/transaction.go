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

	CurrentMonthSummary(c context.Context, query SummaryQueryInput) (Summary, error)
}

type Transaction struct {
	ID                string             `json:"id" gorm:"primaryKey"`
	UserID            string             `json:"user_id"` // creator
	Category          string             `json:"category"`
	TransactionType   string             `json:"transaction_type"` // e.g. "INCOME", "EXPENSE"
	Description       string             `json:"description"`
	SpentAt           time.Time          `json:"spent_at"`
	Amount            decimal.Decimal    `json:"amount" gorm:"type:numeric(20,2)"`
	IsShared          bool               `json:"is_shared"`
	TransactionShares []TransactionShare `json:"transaction_shares" gorm:"foreignKey:TransactionID"`
	User              User               `json:"user" gorm:"foreignKey:UserID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type TransactionShare struct {
	ID            string          `json:"id" gorm:"primaryKey"`
	TransactionID string          `json:"transaction_id"`
	UserID        string          `json:"user_id"`
	Percentage    decimal.Decimal `json:"percentage"` // example: 50.00
	Amount        decimal.Decimal `json:"amount" gorm:"type:numeric(20,2)"`
	User          User            `json:"user" gorm:"foreignKey:UserID"`
	Transaction   Transaction     `json:"-" gorm:"foreignKey:TransactionID"` // avoid recursion
}

type TransactionQueryInput struct {
	Keyword string `query:"keyword"`
	UserID  string `query:"user_id"`
	PaginatedRequest
}

type SummaryQueryInput struct {
	UserID    string `query:"user_id"`
	StartDate string `query:"start_date"` // format: "2006-01-02"
	EndDate   string `query:"end_date"`   // format: "2006-01-02"
}

type Summary struct {
	TotalExpense decimal.Decimal `json:"total_expense"`
	TotalSplited SplitedSummary  `json:"total_splited"`
}

type SplitedSummary struct {
	Me     decimal.Decimal `json:"me"`
	Shared decimal.Decimal `json:"shared"`
}
