package main

import (
	"time"
	"trader/cmd/lib/coinbase"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const server string = "wss://ws-feed.exchange.coinbase.com"

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
	log.Info().Msgf("Starting up, dialing Coinbase at: %s", server)

	c, _, err := websocket.DefaultDialer.Dial(server, nil)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer c.Close()

	log.Info().Msg("Sending subscribe message")

	if err = subscribe(c); err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Info().Msg("Subscribed to coinbase ticker")

	// Read ticket messages from the websocket connection
	messages := make(chan coinbase.TickerMessage, 8)

	go func() {
		log.Info().Msg("Starting to fetch ticker info")

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
		log.Info().Str("product_id", message.ProductId).Str("price", message.Price).Send()
	}

	// In case there are no messages to process, make the main thread wait
	select {}
}
