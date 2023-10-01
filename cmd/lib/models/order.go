package models

import "time"

const (
	BUY_ORDER = iota
	SELL_ORDER
	HOLD_ORDER
)

type Order struct {
	Timestamp time.Time `gorm:"primaryKey"`
	Type      int
}
