package mapapi

import "context"

type Client struct{}

func NewClient() *Client { return &Client{} }

func (c *Client) Distance(_ context.Context, fromLat, fromLng, toLat, toLng float64) (float64, error) {
dx := fromLat - toLat
dy := fromLng - toLng
if dx < 0 { dx = -dx }
if dy < 0 { dy = -dy }
return (dx + dy) * 111, nil
}
