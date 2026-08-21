package repositories

import "errors"

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderStockUnavailable = errors.New("order stock unavailable")
	ErrCartChanged           = errors.New("cart changed during checkout")
)
