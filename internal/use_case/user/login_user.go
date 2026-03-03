package user

import (
	"context"

	"github.com/ogiovannyoliveira/go-bid/internal/validator"
)

type LoginUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req LoginUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.Matches(req.Email, validator.EmailRX), "email", "Email must be valid")
	eval.CheckField(validator.NotBlank(req.Password), "password", "This field cannot be empty")

	return eval
}
