package model

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WalletRepository interface {
	Create(c context.Context, wallet *Wallet) error
	FindAll(c context.Context, query WalletQueryInput) ([]Wallet, int64, error)
	FindByID(c context.Context, id int64) (Wallet, error)
	Update(c context.Context, id int64, wallet Wallet) error
	Delete(c context.Context, id int64) error
	Option(c context.Context, userID string) ([]Wallet, error)
}

type Wallet struct {
	ID        int64           `json:"id"`
	UserID    string          `json:"user_id"`
	Name      string          `json:"name"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `json:"deleted_at"`
}

type WalletInput struct {
	UserID  string          `json:"user_id"`
	Name    string          `json:"name"`
	Balance decimal.Decimal `json:"balance"`
}

type WalletQueryInput struct {
	Keyword string `query:"keyword"`
	UserID  string `query:"user_id"`
	PaginatedRequest
}

func (w *Wallet) InitiateTransactionBalance() Transaction {
	return Transaction{
		UserID:          w.UserID,
		WalletID:        w.ID,
		Amount:          w.Balance,
		Category:        CategoryOpname,
		TransactionType: TransactionTypeIncome,
		Description:     "Initial balance",
		SpentAt:         time.Now(),
	}
}
