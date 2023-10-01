package main

import (
	"fmt"
	"os"
	"time"
	"trader/cmd/lib/coinbase"
	"trader/cmd/lib/models"

	"github.com/cinar/indicator"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var postgresPassword string = os.Getenv("POSTGRES_PASSWORD")

const postgresHost = "postgres"
const postgresPort = 5432
const postgresUsername = "trader"
const postgresDb = "trader"

func main() {
	log.Info().Msg("Starting up")

	db, err := gorm.Open(
		postgres.Open(fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",

			postgresHost,
			postgresUsername,
			postgresPassword,
			postgresDb,
			postgresPort,
		)),
		&gorm.Config{},
	)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	db.AutoMigrate(&models.Candle{})

	log.Info().Msgf("Connected to postgres!")
	log.Info().
		Str("host", postgresHost).
		Int("port", postgresPort).
		Str("username", postgresUsername).
		Str("db", postgresDb).
		Send()

	// Looking at the past n periods
	// This calculates the start time for the API call and it's the size
	// of the outside array in the [][]float64 response
	const requestedWindow = 15

	// Read minute by minute rates and calculate momentum
	for {
		now := time.Now()

		// Get 1m granular candle data for the past minute
		// Returns one bucket
		rawCandles, err := coinbase.GetHistoricCandles(
			"BTC-USD",
			coinbase.GRANULARITY_1M,
			now.Add(-requestedWindow*time.Minute),
			now,
		)
		if err != nil {
			log.Error().Err(err).Send()
		}

		// API makes best effort to return all periods, but not guaranteed at all
		window := len(rawCandles)
		log.Info().Int("window", window).Send()

		// Convert [][]float64 to []Candle
		// Lets us get nice types instead of messing around with indexes,
		// but means we need to loop a second time when calculating indicators
		candles := make([]*models.Candle, window)
		for i := 0; i < window; i++ {
			candles[i] = models.NewCandle(rawCandles[i])
		}

		// Candles are guaranteed sorted by recent -> oldest
		latestCandle := candles[0]

		// Calculate momentum indicators and insert into postgres
		// Indicators look at the past n periods and calculate a new
		// value for this period

		lows := make([]float64, window)
		highs := make([]float64, window)
		closes := make([]float64, window)

		for i := 0; i < window; i++ {
			// Need to reverse the lists so the most recent is at the end
			lows[i] = candles[window-i-1].Low
			highs[i] = candles[window-i-1].High
			closes[i] = candles[window-i-1].Close
		}

		// The indicator value for this period should be the last item in the slice

		// Calculate RSI over n actual periods
		_, rsi := indicator.RsiPeriod(window, closes)

		// Calculate Stochastic Oscillator over past 14m window
		stoc_k, stoc_d := indicator.StochasticOscillator(highs, lows, closes)

		// Update latest candle with calculated indicators to store in the db
		latestCandle.RSI = rsi[len(rsi)-1]
		latestCandle.STOC_K = stoc_k[len(stoc_k)-1]
		latestCandle.STOC_D = stoc_d[len(stoc_d)-1]
		log.Info().
			Float64("close", latestCandle.Close).
			Float64("rsi", latestCandle.RSI).
			Float64("stoc_k", latestCandle.STOC_K).
			Float64("stoc_d", latestCandle.STOC_D).
			Send()

		db.Create(latestCandle)

		// The interval between ticks is long enough to just do our
		// buy and sell orders right here

		time.Sleep(1 * time.Minute)
	}
}
