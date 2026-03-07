package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/ogiovannyoliveira/go-bid/internal/jsonutils"
	"github.com/ogiovannyoliveira/go-bid/internal/services"
	"github.com/ogiovannyoliveira/go-bid/internal/use_case/product"
)

func (api *Api) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[product.CreateProductReq](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	userId, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "Unexpected error, try again later",
		})
		return
	}

	productId, err := api.ProductService.CreateProduct(
		r.Context(),
		userId,
		data.ProductName,
		data.Description,
		data.BasePrice,
		data.AuctionEnd,
	)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
			"error": "Failed to create product auction",
		})
		return
	}

	ctx, _ := context.WithDeadline(context.Background(), data.AuctionEnd)
	auctionRoom := services.NewAuctionRoom(ctx, productId, api.BidsService)

	go auctionRoom.Run()

	api.AuctionLobby.Lock()
	api.AuctionLobby.Rooms[productId] = auctionRoom
	api.AuctionLobby.Unlock()

	jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"message": "Auction has started successfully",
		"id":      productId,
	})
}
