package api

import (
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/ogiovannyoliveira/go-bid/internal/services"
)

type Api struct {
	Router         *chi.Mux
	Sessions       *scs.SessionManager
	UserService    services.UserService
	ProductService services.ProductService
	BidsService    services.BidsService
	WSUpgrader     websocket.Upgrader
	AuctionLobby   services.AuctionLobby
}
