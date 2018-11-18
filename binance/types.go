package main
import ()
type binanceOrderBook struct {
	Bids [][]interface{} `json:"bids"`
	Asks [][]interface{} `json:"asks"` 
	LastUpdated int `json:"lastUpdateId"`
	LocalTime int64 `json:"localTime"`
}

type exchangeInfo struct{
	Symbols []struct {
		Symbol string `json:"symbol"`
		Status string `json:"status"` //should be TRADING
		Filters []struct {
			FilterType string `json:"filterType"`
			StepSize string `json:"stepSize"`
		} `json:"filters"`
	} `json:"symbols"`
}

type binanceOrderBookDiff struct {
	ExchangeTime int64 `json:"E" bson:"ExchangeTime"` 
	EventType string `json:"e"`
	Symbol string `json:"s" bson:"Symbol"`
	FirstUpdate	int `json:"U" bson:"FirstUpdate"`
	LastUpdate int `json:"u" bson:"LastUpdate"`
	Bids [][]interface{} `json:"b,string" bson:"Bids"`
	Asks [][]interface{} `json:"a,string" bson:"Asks"`
	LocalTime int64 `bson:"LocalTime"`
}
type binanceCandlestick struct {
	ExchangeTime int64 `json:"E" bson:"ExchangeTime"` 
	EventType string `json:"e"`
	LocalTime int64 `bson:"LocalTime"`
	Symbol string `json:"s" bson:"Symbol"`
	Kline struct {
		StartTime int64 `json:"t"`
		EndTime int64 `json:"T"`
		Interval string `json:"i"`
		OpenPrice float64 `json:"o,string"`
		ClosePrice float64 `json:"c,string"`
		HighPrice float64 `json:"h,string"`
		LowPrice float64 `json:"l,string"`
		VolumeBase float64 `json:"v,string"`	//this how much of it there is
		NumberOfTrades int64 `json:"n"`
		Closed bool `json:"x"`
		BackingAssetVolume float64 `json:"q,string"`
		BuyVolume float64 `json:"V,string"`
		BuyBackingVolume float64 `json:"Q,string"` //ie eth in trxeth
		LastTradeId int64 `json:"L"`
	} `json:"k"`
}