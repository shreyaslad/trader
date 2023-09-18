package main

import (
	"context"
	"os"
	"strconv"
	"time"
	"trader/cmd/lib/coinbase"

	"github.com/gorilla/websocket"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/rs/zerolog/log"
)

const influxEndpoint = "http://influx:8086"
const coinbaseEndpoint = "wss://ws-feed.exchange.coinbase.com"

var influxToken string = os.Getenv("INFLUXDB_TOKEN")

const influxOrg = "testorg"
const influxBucket = "ticker"

func subscribe(c *websocket.Conn) error {
	subscribeMessage := coinbase.SubscribeMessage{
		Type:       "subscribe",
		ProductIds: []string{"BTC-USD"},
		Channels: []coinbase.Channel{
			{
				Name: "ticker",
			},
		},
	}

	err := c.WriteJSON(subscribeMessage)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.Info().Msg("Starting up")

	influxClient := influxdb2.NewClient(influxEndpoint, influxToken)
	influxWriter := influxClient.WriteAPIBlocking(influxOrg, influxBucket)

	log.Info().Msgf("Connected to influx at: %s", influxEndpoint)

	c, _, err := websocket.DefaultDialer.Dial(coinbaseEndpoint, nil)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer c.Close()

	log.Info().Msgf("Dialed Coinbase at: %s", coinbaseEndpoint)

	if err = subscribe(c); err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Info().Msg("Subscribed to coinbase ticker")

	// Read ticket messages from the websocket connection
	messages := make(chan coinbase.TickerMessage, 8)

	go func() {
		log.Info().Msg("Fetching ticker info. No individual ticker logs will be shown")

		defer close(messages)
		for {
			message := coinbase.TickerMessage{}
			if err := c.ReadJSON(&message); err != nil {
				log.Error().Err(err).Send()
				break
			}

			messages <- message

			// Hit the Coinbase ratelimit of 8 messages / second
			time.Sleep(125 * time.Millisecond)
		}
	}()

	for message := range messages {
		if message.Price == "" {
			continue
		}

		floatPrice, err := strconv.ParseFloat(message.Price, 64)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		// "layout" argument for time.Parse is specific value
		// https://stackoverflow.com/a/25845833
		tickTime, err := time.Parse("2006-01-02T15:04:05.000000Z", message.Time)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		tickPoint := write.NewPoint(
			"price",
			map[string]string{"product_id": message.ProductId},
			map[string]interface{}{"price": floatPrice},
			tickTime,
		)

		influxWriter.WritePoint(context.Background(), tickPoint)
		// log.Info().Str("product_id", message.ProductId).Str("price", message.Price).Send()
	}

	// In case there are no messages to process, make the main thread wait
	select {}
}
