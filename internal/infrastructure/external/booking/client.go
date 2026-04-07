package booking

import "context"

type Client struct{}

type BookingRequest struct {
ProductID string
Quantity  int
}

type BookingResponse struct {
OrderID string
Status  string
}

func NewClient() *Client { return &Client{} }

func (c *Client) CreateOrder(_ context.Context, req BookingRequest) (BookingResponse, error) {
return BookingResponse{OrderID: "order-" + req.ProductID, Status: "confirmed"}, nil
}
