package models

import "time"

type Candle struct {
	Timestamp time.Time `gorm:"primaryKey"`
	Low       float64
	High      float64
	Open      float64
	Close     float64

	RSI    float64
	STOC_K float64
	STOC_D float64
}

func NewCandle(rawCandle []float64) *Candle {
	// Indexes (this is great, thanks coinbase)
	// timestamp: 0
	// low: 1
	// high: 2
	// open: 3
	// close: 4

	return &Candle{
		Timestamp: time.Unix(int64(rawCandle[0]), 0),
		Low:       rawCandle[1],
		High:      rawCandle[2],
		Open:      rawCandle[3],
		Close:     rawCandle[4],
	}
}
