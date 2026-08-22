package main

import (
	"context"
	"errors"
	"fmt"
)

type OrderRequest struct {
	Number   int
	Item     string
	Quantity uint8
}

func (o *OrderRequest) Validate(ctx context.Context) error {
	if o.Item == "" {
		return errors.New("cannot order an empty item")
	}
	if o.Quantity == 0 {
		return errors.New("cannot order zero items")
	}
	return nil
}

type OnlineStore struct{}

func (o *OnlineStore) Order(ctx context.Context, r *OrderRequest) (bool, error) {
	if r.Number != 1 {
		return false, errors.New("order number extracted from URL path must be 1")
	}
	if r.Item == "" {
		return false, errors.New("cannot order an empty item")
	}
	if r.Quantity == 0 {
		return false, errors.New("cannot order zero items")
	}
	return true, nil
}

func (o *OnlineStore) GetPrice(ctx context.Context, item string) (float64, error) {
	switch item {
	case "shirt":
		return 10.99, nil
	case "pants":
		return 9.99, nil
	case "hat":
		return 5.99, nil
	default:
		return 0, fmt.Errorf("unknown item %q", item)
	}
}

func (o *OnlineStore) GetInventory(ctx context.Context) ([]string, error) {
	return []string{"shirt", "pants", "hat"}, nil
}
