package coinbase

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const coinbaseEndpoint = "https://api.exchange.coinbase.com"

const (
	GRANULARITY_1M = 60
	GRANULARITY_5M = 300
)

// Get candles for a given period and granularity from the Coinbase API to perform technical analysis
// THe Coinbase API returns an array of buckets, where each bucket is also an array of values. Indexing
// the array is the best way to properly decode it, so good luck with that
func GetHistoricCandles(productId string, granularity int, start time.Time, end time.Time) ([][]float64, error) {
	candlesEndpoint := fmt.Sprintf(
		"%s/products/%s/candles?granularity=%d&start=%d&end=%d",
		coinbaseEndpoint,

		productId,
		granularity, start.Unix(), end.Unix(),
	)

	client := http.Client{}
	req, err := http.NewRequest("GET", candlesEndpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error().Any("response", res).Msg("failed parsing response body")
		log.Fatal().Err(err).Send()
	}

	if res.StatusCode != http.StatusOK {
		log.Error().Err(errors.New("failed to get 200 OK from coinbase candle endpoint")).Send()
		log.Fatal().Msg(string(resBody))
	}

	var data [][]float64
	if err = json.Unmarshal(resBody, &data); err != nil {
		return nil, err
	}

	return data, nil
}
