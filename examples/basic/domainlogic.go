package main

import (
	"context"
	"errors"
	"fmt"

	"log/slog"
)

type OrderRequest struct {
	Item     string
	Quantity uint8
}

func (o *OrderRequest) Validate() error {
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

func (o *OnlineStore) Record(ctx context.Context, r *OrderRequest) error {
	slog.InfoContext(
		ctx,
		"received order",
		slog.String("item", r.Item),
		slog.Any("quantity", r.Quantity),
	)
	return nil
}
