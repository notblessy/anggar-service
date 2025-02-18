package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/notblessy/anggar-service/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

func (u *userRepository) verifyToken(ctx context.Context, code, requestOrigin string) (model.GoogleAuthInfo, error) {
	logger := logrus.WithFields(logrus.Fields{
		"code":   code,
		"origin": requestOrigin,
	})

	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  requestOrigin,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"openid", "https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return model.GoogleAuthInfo{}, fmt.Errorf("code exchange wrong: %s", err.Error())
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		logger.Error(model.ErrGoogleNoIdToken)
		return model.GoogleAuthInfo{}, model.ErrGoogleNoIdToken
	}

	payload, err := idtoken.Validate(context.Background(), idToken, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		logger.Errorf("Error validating token: %v", err)
		return model.GoogleAuthInfo{}, fmt.Errorf("id token validation failed: %w", model.ErrInvalidAuthClaim)
	}

	return u.claimAuth(payload)
}

func (u *userRepository) claimAuth(payload *idtoken.Payload) (model.GoogleAuthInfo, error) {

	email := payload.Claims["email"].(string)
	if email == "" {
		return model.GoogleAuthInfo{}, fmt.Errorf("email: %w", model.ErrInvalidAuthClaim)
	}

	name := payload.Claims["name"].(string)
	if name == "" {
		return model.GoogleAuthInfo{}, fmt.Errorf("name: %w", model.ErrInvalidAuthClaim)
	}

	picture := payload.Claims["picture"].(string)

	return model.GoogleAuthInfo{
		Email:   email,
		Name:    name,
		Picture: picture,
	}, nil
}
