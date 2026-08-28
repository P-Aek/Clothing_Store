package repositories

import "errors"

var (
	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderNotOwned           = errors.New("order does not belong to user")
	ErrOrderAlreadyCancelled   = errors.New("order already cancelled")
	ErrOrderCannotBeCancelled  = errors.New("order cannot be cancelled in its current status")
	ErrOrderStockUnavailable   = errors.New("order stock unavailable")
	ErrOrderStockRestoreFailed = errors.New("order stock restoration failed")
	ErrCartChanged             = errors.New("cart changed during checkout")
)
