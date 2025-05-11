package model

import (
	"context"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WalletRepository interface {
	Create(c context.Context, wallet *Wallet) error
	FindAll(c context.Context, query WalletQueryInput) ([]Wallet, int64, error)
	FindByID(c context.Context, id string) (Wallet, error)
	Update(c context.Context, id string, wallet Wallet) error
	Delete(c context.Context, id string) error
	Option(c context.Context, userID string) ([]Wallet, error)
}

type Wallet struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Name      string          `json:"name"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `json:"deleted_at"`
	Owner     User            `json:"owner" gorm:"foreignKey:UserID;->"`
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
		ID:              ulid.Make().String(),
		UserID:          w.UserID,
		Amount:          w.Balance,
		Category:        CategoryOpname,
		TransactionType: TransactionTypeIncome,
		Description:     "Initial balance",
		SpentAt:         time.Now(),
	}
}
