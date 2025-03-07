package repository

import (
	"context"
	"time"

	"github.com/notblessy/anggar-service/model"
	"github.com/notblessy/anggar-service/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type scopeRepository struct {
	db *gorm.DB
}

// NewScopeRepository :nodoc:
func NewScopeRepository(db *gorm.DB) model.ScopeRepository {
	return &scopeRepository{db}
}

func (r *scopeRepository) Create(c context.Context, scope model.ScopeInput) (model.Scope, error) {
	logger := logrus.WithField("scope", utils.Dump(scope))

	scopeCreate := scope.ToScope()

	tx := r.db.WithContext(c).Begin()

	if err := tx.Create(&scopeCreate).Error; err != nil {
		logger.Error(err)
		tx.Rollback()
		return model.Scope{}, err
	}

	var scopeCategories []model.ScopeCategory

	for _, category := range scope.CategoryIDs {
		scopeCategories = append(scopeCategories, model.ScopeCategory{
			ScopeID:  scopeCreate.ID,
			Category: category,
		})
	}

	if err := tx.Create(&scopeCategories).Error; err != nil {
		logger.Error(err)
		tx.Rollback()
		return model.Scope{}, err
	}

	return scopeCreate, nil
}

func (r *scopeRepository) FindAll(c context.Context, query model.ScopeQueryInput) ([]model.Scope, int64, error) {
	logger := logrus.WithField("query", utils.Dump(query))

	var scopes []model.Scope

	qb := r.db.WithContext(c).Where("user_id = ?", query.UserID)

	if query.Keyword != "" {
		qb = qb.Where("name ILIKE ?", "%"+query.Keyword+"%")
	}

	var total int64
	if err := qb.Model(&model.Scope{}).Count(&total).Error; err != nil {
		logger.Error(err)
		return nil, 0, err
	}

	if err := qb.Scopes(query.Paginated()).Order(query.Sorted()).Find(&scopes).Error; err != nil {
		logger.Error(err)
		return nil, 0, err
	}

	return scopes, total, nil
}

func (r *scopeRepository) FindByID(c context.Context, id int64) (model.Scope, error) {
	logger := logrus.WithField("id", id)

	var scope model.Scope

	if err := r.db.First(&scope, id).Error; err != nil {
		logger.Error(err)
		return model.Scope{}, err
	}

	return scope, nil
}

func (r *scopeRepository) Update(c context.Context, id int64, scope model.Scope) error {
	logger := logrus.WithField("id", id).WithField("scope", utils.Dump(scope))

	updatedBudgt := model.Scope{
		ID:        id,
		UserID:    scope.UserID,
		Name:      scope.Name,
		CreatedAt: scope.CreatedAt,
		UpdatedAt: time.Now(),
	}

	trx := r.db.WithContext(c).Begin()

	if err := trx.Model(&model.Scope{}).Where("id = ?", id).Updates(updatedBudgt).Error; err != nil {
		logger.Error(err)
		trx.Rollback()
		return err
	}

	if len(scope.ScopeCategories) > 0 {
		if err := trx.Where("scope_id = ?", id).Delete(&model.ScopeCategory{}).Error; err != nil {
			logger.Error(err)
			trx.Rollback()
			return err
		}

		for i := range scope.ScopeCategories {
			scope.ScopeCategories[i].ScopeID = id
		}

		if err := trx.Create(&scope.ScopeCategories).Error; err != nil {
			logger.Error(err)
			trx.Rollback()
			return err
		}
	} else {
		if err := trx.Where("scope_id = ?", id).Delete(&model.ScopeCategory{}).Error; err != nil {
			logger.Error(err)
			trx.Rollback()
			return err
		}
	}

	return nil
}

func (r *scopeRepository) Delete(c context.Context, id int64) error {
	logger := logrus.WithField("id", id)

	if err := r.db.Delete(&model.Scope{}, id).Error; err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (r *scopeRepository) FindOverviews(c context.Context, userID string) ([]model.ScopeOverview, error) {
	logger := logrus.WithField("userID", userID)

	var overviews []model.ScopeOverview

	if err := r.db.
		Table("scopes").
		Select(`
			scope.id, 
			scope.user_id, 
			scope.name,
			scope.created_at, 
			scope.updated_at, 
			scope.deleted_at, 
			COALESCE(SUM(transactions.amount), 0) AS total_amount_transaction
		`).
		Joins("LEFT JOIN scope_categories ON scope.id = scope_categories.scope_id").
		Joins("LEFT JOIN transactions ON transactions.category = scope_categories.category").
		Where("scope.user_id = ?", userID).
		Group("scope.id").
		Scan(&overviews).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	return overviews, nil
}
