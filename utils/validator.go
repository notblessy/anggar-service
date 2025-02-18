package utils

import "github.com/go-playground/validator"

type Ghost struct {
	Validator *validator.Validate
}

func (g *Ghost) Validate(i interface{}) error {
	if err := g.Validator.Struct(i); err != nil {
		return err
	}

	return nil
}
