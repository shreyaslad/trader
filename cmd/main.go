package main

import (
	"fmt"
	"os"
	"time"
	"trader/cmd/lib/coinbase"

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

	_, err := gorm.Open(
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

	log.Info().Msgf("Connected to postgres!")
	log.Info().
		Str("host", postgresHost).
		Int("port", postgresPort).
		Str("username", postgresUsername).
		Str("db", postgresDb).
		Send()

	// Read minute by minute rates and calculate momentum
	for {
		now := time.Now()

		// Get 1m granular candle data for the past minute
		// Returns one bucket
		candles, err := coinbase.GetHistoricCandles(
			"BTC-USD",
			coinbase.GRANULARITY_1M,
			now.Add(-1*time.Minute),
			now,
		)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		log.Info().Floats64("1m_candle", candles[0]).Send()

		// Calculate momentum indicators and insert into postgres
		// Indicators look at the past 5 periods and calculate a new
		// value for this period

		// TODO: calculate indicators for past 5 buckets, but only insert
		// TODO: most recent one into db

		// The interval between ticks is long enough to just do our
		// buy and sell orders right here

		time.Sleep(1 * time.Minute)
	}
}
