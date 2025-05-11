package repository

import (
	"context"
	"fmt"
	"os"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/notblessy/anggar-service/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/utils"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository :nodoc:
func NewUserRepository(d *gorm.DB) model.UserRepository {
	return &userRepository{
		db: d,
	}
}

func (a *userRepository) Authenticate(ctx context.Context, code, requestOrigin string) (model.User, error) {
	logger := logrus.WithFields(logrus.Fields{
		"code":           code,
		"request_origin": requestOrigin,
	})

	auth, err := a.verifyToken(ctx, code, requestOrigin)
	if err != nil {
		logger.Errorf("Error verifying token: %v", err)
		return model.User{}, err
	}

	id, err := gonanoid.New()
	if err != nil {
		logger.Errorf("Error generating id: %v", err)
		return model.User{}, err
	}

	validEmails := os.Getenv("VALID_EMAILS")
	if validEmails != "" {
		validEmailList := strings.Split(validEmails, ",")
		if !utils.Contains(validEmailList, auth.Email) {
			logger.Errorf("Email %s is not in the list of valid emails", auth.Email)
			return model.User{}, fmt.Errorf("email %s is not in the list of valid emails", auth.Email)
		}
	}

	var authUser model.User

	err = a.db.Where("email = ?", auth.Email).First(&authUser).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.Errorf("Error querying user: %v", err)
		return model.User{}, err
	}

	if err == gorm.ErrRecordNotFound {
		authUser = model.User{
			ID:      id,
			Name:    auth.Name,
			Role:    "USER",
			Email:   auth.Email,
			Picture: auth.Picture,
		}

		err = a.db.Create(&authUser).Error
		if err != nil {
			logger.Errorf("Error creating user: %v", err)
			return model.User{}, err
		}
	}

	return authUser, nil
}

func (a *userRepository) FindByID(ctx context.Context, id string) (model.User, error) {
	logger := logrus.WithField("id", id)

	var user model.User
	err := a.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		logger.Errorf("Error querying user: %v", err)
		return model.User{}, err
	}

	user.OmitPassword()

	return user, nil
}

func (a *userRepository) FindOptions(ctx context.Context) ([]model.UserOption, error) {
	logger := logrus.WithField("context", ctx)

	var options []model.UserOption

	err := a.db.Table("users").Order("name ASC").Find(&options).Error
	if err != nil {
		logger.Errorf("Error querying user options: %v", err)
		return nil, err
	}

	return options, nil
}
