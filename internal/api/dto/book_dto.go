package dto

import "github.com/go-playground/validator/v10"

type CreateBookRequest struct {
	Title  string `json:"title" validate:"required,min=1,max=200"`
	Author string `json:"author" validate:"required,min=1,max=200"`
	Year   int    `json:"year" validate:"required,min=0,max=2026"`
}

type UpdateBookRequest struct {
	Title  *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Author *string `json:"author,omitempty" validate:"omitempty,min=1,max=200"`
	Year   *int    `json:"year,omitempty" validate:"omitempty,min=0,max=2026"`
}

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func (r *CreateBookRequest) Validate() error {
	return validate.Struct(r)
}

func (r *UpdateBookRequest) Validate() error {
	return validate.Struct(r)
}

