package user

import (
	"github.com/go-playground/validator/v10"
)

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name"  validate:"required"`
	Email     string `json:"email"      validate:"required,email"`
	Password  string `json:"password"   validate:"required,min=6"`
}

var validate = validator.New()

func New(firstName, lastName, email, password string) (*User, error) {
	u := &User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  password,
	}
	if err := validate.Struct(u); err != nil {
		return nil, err
	}
	return u, nil
}
