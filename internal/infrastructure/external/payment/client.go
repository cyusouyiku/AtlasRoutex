package payment

import "context"

type Client struct{}

type ChargeResult struct {
TransactionID string
Status        string
}

func NewClient() *Client { return &Client{} }

func (c *Client) Charge(_ context.Context, orderID string, amount float64) (ChargeResult, error) {
return ChargeResult{TransactionID: "txn-" + orderID, Status: "paid"}, nil
}
