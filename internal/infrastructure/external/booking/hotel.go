package booking

import "context"

type HotelClient struct { base *Client }

func NewHotelClient(base *Client) *HotelClient { return &HotelClient{base: base} }

func (h *HotelClient) BookRoom(ctx context.Context, hotelID string, nights int) (BookingResponse, error) {
return h.base.CreateOrder(ctx, BookingRequest{ProductID: hotelID, Quantity: nights})
}
