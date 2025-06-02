package repository

import (
	"context"

	"github.com/notblessy/anggar-service/model"
	"github.com/notblessy/anggar-service/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) model.TransactionRepository {
	return &transactionRepository{db}
}

func (r *transactionRepository) Create(c context.Context, transaction *model.Transaction) error {
	logger := logrus.WithField("transaction", utils.Dump(transaction))

	if err := r.db.Create(transaction).Error; err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (r *transactionRepository) FindAll(c context.Context, query model.TransactionQueryInput) ([]model.Transaction, int64, error) {
	logger := logrus.WithField("query", utils.Dump(query))

	var transactions []model.Transaction

	qb := r.db.WithContext(c).Preload("User").Preload("TransactionShares.User")

	if query.UserID != "" {
		qb.Where("user_id = ?", query.UserID)
	}

	if query.Keyword != "" {
		qb = qb.Where("name ILIKE ?", "%"+query.Keyword+"%")
	}

	var total int64
	if err := qb.Model(&model.Transaction{}).Count(&total).Error; err != nil {
		logger.Error(err)
		return nil, 0, err
	}

	if err := qb.Scopes(query.Paginated()).Order(query.Sorted()).Find(&transactions).Error; err != nil {
		logger.Error(err)
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *transactionRepository) FindByID(c context.Context, id int64) (model.Transaction, error) {
	logger := logrus.WithField("id", id)

	var transaction model.Transaction
	if err := r.db.WithContext(c).Where("id = ?", id).First(&transaction).Error; err != nil {
		logger.Error(err)
		return model.Transaction{}, err
	}

	return transaction, nil
}

func (r *transactionRepository) Update(c context.Context, id int64, transaction model.Transaction) error {
	logger := logrus.WithField("transaction", utils.Dump(transaction))

	if err := r.db.WithContext(c).Model(&model.Transaction{}).Where("id = ?", id).Updates(&transaction).Error; err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (r *transactionRepository) Delete(c context.Context, id int64) error {
	logger := logrus.WithField("id", id)

	if err := r.db.WithContext(c).Delete(&model.Transaction{}, id).Error; err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (r *transactionRepository) CurrentMonthSummary(c context.Context, query model.SummaryQueryInput) (model.Summary, error) {
	logger := logrus.WithField("query", utils.Dump(query))

	var summary model.Summary

	if err := r.db.WithContext(c).
		Model(&model.Transaction{}).
		Select("COALESCE(SUM(amount), 0) AS total_expense").
		Where("user_id = ? AND transaction_type = ?", query.UserID, model.TransactionTypeIncome).
		Where("DATE(spent_at) BETWEEN ? AND ?", query.StartDate, query.EndDate).
		Scan(&summary.TotalExpense).Error; err != nil {
		logger.Error(err)
		return model.Summary{}, err
	}

	transactionIds := []string{}

	if err := r.db.WithContext(c).
		Model(&model.Transaction{}).
		Select("id").
		Where("user_id = ? AND transaction_type = ?", query.UserID, model.TransactionTypeExpense).
		Where("DATE(spent_at) BETWEEN ? AND ?", query.StartDate, query.EndDate).
		Find(&transactionIds).Error; err != nil {
		logger.Error(err)
		return model.Summary{}, err
	}

	if err := r.db.WithContext(c).
		Model(&model.TransactionShare{}).
		Select("COALESCE(SUM(amount), 0) AS me").
		Where("transaction_id IN ?", transactionIds).
		Where("user_id = ?", query.UserID).
		Scan(&summary.TotalSplited.Me).Error; err != nil {
		logger.Error(err)
		return model.Summary{}, err
	}

	if err := r.db.WithContext(c).
		Model(&model.TransactionShare{}).
		Select("COALESCE(SUM(amount), 0) AS shared").
		Where("transaction_id IN ?", transactionIds).
		Where("user_id <> ?", query.UserID).
		Scan(&summary.TotalSplited.Shared).Error; err != nil {
		logger.Error(err)
		return model.Summary{}, err
	}

	return summary, nil

}
