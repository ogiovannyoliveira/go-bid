package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ogiovannyoliveira/go-bid/internal/store/pgstore"
)

type ProductService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewProductService(pool *pgxpool.Pool) ProductService {
	return ProductService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (ps *ProductService) CreateProduct(
	ctx context.Context,
	sellerId uuid.UUID,
	productName,
	description string,
	basePrice float64,
	auctionEnd time.Time,
) (uuid.UUID, error) {
	args := pgstore.CreateProductParams{
		SellerID:    sellerId,
		ProductName: productName,
		Description: description,
		BasePrice:   basePrice,
		AuctionEnd:  auctionEnd,
	}
	id, err := ps.queries.CreateProduct(ctx, args)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

var ErrProductNotFound = errors.New("Product not found")

func (ps *ProductService) GetProductById(ctx context.Context, id uuid.UUID) (pgstore.Product, error) {
	product, err := ps.queries.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Product{}, ErrProductNotFound
		}
		return pgstore.Product{}, err
	}

	return product, nil
}
