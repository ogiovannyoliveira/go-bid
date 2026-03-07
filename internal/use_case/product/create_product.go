package product

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ogiovannyoliveira/go-bid/internal/validator"
)

type CreateProductReq struct {
	SellerID    uuid.UUID `json:"seller_id"`
	ProductName string    `json:"product_name"`
	Description string    `json:"description"`
	BasePrice   float64   `json:"base_price"`
	AuctionEnd  time.Time `json:"auction_end"`
}

const minAuctionDuration = 2 * time.Hour

func (req CreateProductReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.ProductName), "product_name", "This field cannot be empty")
	eval.CheckField(validator.MinChars(req.Description, 10) && validator.MaxChars(req.Description, 255), "description", "This field must have a length between 10 and 255")
	eval.CheckField(req.BasePrice > 0, "base_price", "This field must be greater than zero")

	eval.CheckField(time.Until(req.AuctionEnd) >= minAuctionDuration, "auction_end", "Must be at least 2 hours duration")

	return eval
}
