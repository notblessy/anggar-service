package repository

import (
	"context"

	"github.com/notblessy/anggar-service/model"
	"github.com/notblessy/anggar-service/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) model.WalletRepository {
	return &walletRepository{db}
}

func (r *walletRepository) Create(ctx context.Context, wallet *model.Wallet) error {
	logger := logrus.WithField("wallet", utils.Dump(wallet))

	trx := r.db.WithContext(ctx).Begin()

	if err := trx.Create(wallet).Error; err != nil {
		logger.Error(err)
		trx.Rollback()
		return err
	}

	transaction := wallet.InitiateTransactionBalance()
	if err := trx.Create(&transaction).Error; err != nil {
		logger.Error(err)
		trx.Rollback()
	}

	trx.Commit()
	return nil
}

func (r *walletRepository) FindAll(ctx context.Context, query model.WalletQueryInput) ([]model.Wallet, int64, error) {
	logger := logrus.WithField("query", utils.Dump(query))

	var wallets []model.Wallet

	qb := r.db.WithContext(ctx).Preload("Owner")

	if query.UserID != "" {
		qb.Where("user_id = ?", query.UserID)
	}

	if query.Keyword != "" {
		qb = qb.Where("name ILIKE ?", "%"+query.Keyword+"%")
	}

	var total int64
	if err := qb.Model(&model.Wallet{}).Count(&total).Error; err != nil {
		logger.Error(err)
		return nil, 0, err
	}

	if err := qb.Scopes(query.Paginated()).Order(query.Sorted()).Find(&wallets).Error; err != nil {
		logger.Error(err)
		return nil, 0, err
	}

	return wallets, total, nil
}

func (r *walletRepository) FindByID(ctx context.Context, id string) (model.Wallet, error) {
	logger := logrus.WithField("id", id)

	var wallet model.Wallet

	if err := r.db.First(&wallet, "id = ?", id).Error; err != nil {
		logger.Error(err)
		return model.Wallet{}, err
	}

	return wallet, nil
}

func (r *walletRepository) Update(ctx context.Context, id string, wallet model.Wallet) error {
	logger := logrus.WithField("id", id).WithField("wallet", utils.Dump(wallet))

	if err := r.db.Model(&model.Wallet{}).Where("id = ?", id).Updates(wallet).Error; err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (r *walletRepository) Delete(ctx context.Context, id string) error {
	logger := logrus.WithField("id", id)

	tx := r.db.WithContext(ctx)

	err := tx.Where("wallet_id = ?", id).Delete(&model.Transaction{}).Error
	if err != nil {
		logger.Error(err)
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&model.Wallet{}, "id = ?", id).Error; err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (r *walletRepository) Option(ctx context.Context, userID string) ([]model.Wallet, error) {
	logger := logrus.WithField("user_id", userID)

	var wallets []model.Wallet

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&wallets).Error

	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return wallets, nil
}
