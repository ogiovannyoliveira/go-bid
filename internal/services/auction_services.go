package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	// Request
	PlaceBid MessageKind = iota

	// Ok/Success
	SuccessfullyPlacedBid

	// Errors
	FailedToPlaceBid
	InvalidJSON

	// Info
	NewBidPlaced
	AuctionFinished
)

type Message struct {
	Message string      `json:"message,omitempty"`
	Kind    MessageKind `json:"kind"`
	Amount  float64     `json:"amount,omitempty"`
	UserID  uuid.UUID   `json:"user_id,omitempty"`
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

type AuctionRoom struct {
	ID         uuid.UUID
	Context    context.Context
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	Clients    map[uuid.UUID]*Client

	BidsService BidsService
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, bidsService BidsService) *AuctionRoom {
	return &AuctionRoom{
		ID:          id,
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[uuid.UUID]*Client),
		Context:     ctx,
		BidsService: bidsService,
	}
}

type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	Send   chan Message
	UserId uuid.UUID
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userId uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message, 512),
		UserId: userId,
	}
}

const (
	maxMessageSize = 512
	readDeadline   = 60 * time.Second
	pingPeriod     = (readDeadline * 9) / 10
	writeWait      = 10 * time.Second
)

func (c *Client) ReadEventLoop() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(readDeadline))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(readDeadline))
		return nil
	})

	for {
		var m Message
		m.UserID = c.UserId
		err := c.Conn.ReadJSON(&m)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				slog.Error("Unexpected close error", "error", err)
				return
			}

			c.Room.Broadcast <- Message{
				Kind:    InvalidJSON,
				Message: "Invalid JSON was provided",
				UserID:  m.UserID,
			}
			continue
		}

		c.Room.Broadcast <- m
	}
}

func (c *Client) WriteEventLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteJSON(Message{
					Kind:    websocket.CloseMessage,
					Message: "Closing websocket connection",
				})
				return
			}

			if message.Kind == AuctionFinished {
				close(c.Send)
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			err := c.Conn.WriteJSON(message)
			if err != nil {
				c.Room.Unregister <- c
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Unexpected write error", "error", err)
				return
			}
		}
	}
}

func (r *AuctionRoom) Run() {
	slog.Info("Auction has begun", "AuctionID", r.ID)
	defer close(r.Broadcast)

	for {
		select {
		case client := <-r.Register:
			r.registerClient(client)
		case client := <-r.Unregister:
			r.unregisterClient(client)
		case message := <-r.Broadcast:
			r.broadcastMessage(message)
		case <-r.Context.Done():
			slog.Info("Auction has ended", "AuctionID", r.ID)
			for _, client := range r.Clients {
				client.Send <- Message{Kind: AuctionFinished, Message: "Auction has been finished"}
			}
			return
		}
	}
}

func (r *AuctionRoom) registerClient(c *Client) {
	slog.Info("New user connected", "Client", c)
	r.Clients[c.UserId] = c
}

func (r *AuctionRoom) unregisterClient(c *Client) {
	slog.Info("User disconnected", "Client", c)
	delete(r.Clients, c.UserId)
}

func (r *AuctionRoom) broadcastMessage(m Message) {
	slog.Info("New message received", "RoomID", r.ID, "Message", m.Message, "UserID", m.UserID)
	switch m.Kind {
	case PlaceBid:
		bid, err := r.BidsService.PlaceBid(r.Context, r.ID, m.UserID, m.Amount)
		if err != nil {
			if errors.Is(err, ErrBidIsTooLow) {
				if client, ok := r.Clients[m.UserID]; ok {
					client.Send <- Message{Kind: FailedToPlaceBid, Message: ErrBidIsTooLow.Error(), UserID: m.UserID}
				}
			}
			return
		}

		if client, ok := r.Clients[m.UserID]; ok {
			client.Send <- Message{Kind: SuccessfullyPlacedBid, Message: "Your bid was successfully placed", UserID: m.UserID}
		}

		for id, client := range r.Clients {
			if id == m.UserID {
				continue
			}

			newBidMessage := Message{Kind: NewBidPlaced, Message: "A new bid has been placed", Amount: bid.BidAmount, UserID: m.UserID}
			client.Send <- newBidMessage
		}
	case InvalidJSON:
		client, ok := r.Clients[m.UserID]
		if !ok {
			slog.Info("Client not found in map", "UserID", m.UserID)
			return
		}

		client.Send <- m
	}
}
