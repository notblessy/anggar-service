package repository

import (
	"context"

	"github.com/notblessy/anggar-service/model"
	"github.com/notblessy/anggar-service/utils"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type budgetRepository struct {
	db *gorm.DB
}

// NewBudgetRepository :nodoc:
func NewBudgetRepository(db *gorm.DB) model.BudgetRepository {
	return &budgetRepository{db}
}

func (r *budgetRepository) Create(c context.Context, budget *model.Budget) error {
	logger := logrus.WithField("budget", utils.Dump(budget))

	if err := r.db.Create(budget).Error; err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (r *budgetRepository) FindAll(c context.Context, query model.BudgetQueryInput) ([]model.Budget, int64, error) {
	logger := logrus.WithField("query", utils.Dump(query))

	var budgets []model.Budget

	qb := r.db.WithContext(c).Where("user_id = ?", query.UserID)

	if query.Keyword != "" {
		qb = qb.Where("name ILIKE ?", "%"+query.Keyword+"%")
	}

	var total int64
	if err := qb.Model(&model.Budget{}).Count(&total).Error; err != nil {
		logger.Error(err)
		return nil, 0, err
	}

	if err := qb.Scopes(query.Paginated()).Order(query.Sorted()).Find(&budgets).Error; err != nil {
		logger.Error(err)
		return nil, 0, err
	}

	return budgets, total, nil
}

func (r *budgetRepository) FindByID(c context.Context, id int64) (model.Budget, error) {
	logger := logrus.WithField("id", id)

	var budget model.Budget

	if err := r.db.First(&budget, id).Error; err != nil {
		logger.Error(err)
		return model.Budget{}, err
	}

	return budget, nil
}

func (r *budgetRepository) Update(c context.Context, id int64, budget model.Budget) error {
	logger := logrus.WithField("id", id).WithField("budget", utils.Dump(budget))

	if err := r.db.Model(&model.Budget{}).Where("id = ?", id).Updates(budget).Error; err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (r *budgetRepository) Delete(c context.Context, id int64) error {
	logger := logrus.WithField("id", id)

	if err := r.db.Delete(&model.Budget{}, id).Error; err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (r *budgetRepository) FindOverviews(c context.Context, userID string) ([]model.BudgetOverview, error) {
	logger := logrus.WithField("userID", userID)

	var overviews []model.BudgetOverview

	if err := r.db.
		Table("budgets").
		Select(`
			budget.id, 
			budget.user_id, 
			budget.name, 
			budget.amount, 
			budget.start_date, 
			budget.end_date, 
			budget.auto_renew, 
			budget.renewal_period, 
			budget.created_at, 
			budget.updated_at, 
			budget.deleted_at, 
			COALESCE(SUM(transactions.amount), 0) AS total_amount_transaction
		`).
		Joins("LEFT JOIN transactions ON transactions.budget_id = budget.id").
		Where("budget.user_id = ?", userID).
		Group("budget.id").
		Scan(&overviews).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	for i := range overviews {
		overviews[i].Leftout = overviews[i].Amount.Sub(overviews[i].TotalAmountTransaction)
		overviews[i].Progress = overviews[i].TotalAmountTransaction.Div(overviews[i].Amount).Mul(decimal.NewFromInt(100))
	}

	return overviews, nil
}
