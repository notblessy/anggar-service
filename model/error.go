package model

import "errors"

var (
	ErrGoogleNoIdToken  = errors.New("no id_token field in oauth2 token")
	ErrInvalidAuthClaim = errors.New("invalid auth claim")
	ErrRegisterRequired = errors.New("register required")
	ErrForbidden        = errors.New("forbidden request")
)
