package user

import (
	"context"

	"github.com/ogiovannyoliveira/go-bid/internal/validator"
)

type CreateUserReq struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func (req CreateUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.UserName), "user_name", "This field cannot be empty")
	eval.CheckField(validator.MaxChars(req.UserName, 50), "user_name", "User name must have maximum of 50 chars")
	eval.CheckField(validator.NotBlank(req.Email), "email", "This field cannot be empty")
	eval.CheckField(validator.Matches(req.Email, validator.EmailRX), "email", "Must be a valid email")
	eval.CheckField(validator.MinChars(req.Password, 8), "password", "Password must be bigger than 8 chars")
	eval.CheckField(validator.MinChars(req.Bio, 10) && validator.MaxChars(req.Bio, 255), "bio", "This field must have length between 10 and 255")

	return eval
}
