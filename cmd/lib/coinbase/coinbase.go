package coinbase

type Channel struct {
	Name       string   `json:"name"`
	ProductIds []string `json:"product_ids"`
}

type SubscribeMessage struct {
	Type       string    `json:"type"`
	ProductIds []string  `json:"product_ids"`
	Channels   []Channel `json:"channels"`
}

type TickerMessage struct {
	Type        string `json:"type"`
	Sequence    int    `json:"sequence"`
	ProductId   string `json:"product_id"`
	Price       string `json:"price"`
	Open24H     string `json:"open_24h"`
	Volume24H   string `json:"volume_24h"`
	Low24H      string `json:"low_24h"`
	High24H     string `json:"high_24h"`
	Volume30D   string `json:"volume_30d"`
	BestBid     string `json:"best_bid"`
	BestBidSize string `json:"best_bid_size"`
	BestAsk     string `json:"best_ask"`
	BestAskSize string `json:"best_ask_size"`
	Side        string `json:"side"`
	Time        string `json:"time"`
	TradeID     int    `json:"trade_id"`
	LastSize    string `json:"last_size"`
}
