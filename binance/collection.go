package main
import (
	"../util"
	"os"
	"log"
	"github.com/go-redis/redis"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"time"
	"net/http"
	"encoding/json"
	"strings"
)
type dayStats struct {
	Volume string `json:"quotevolume"`
	Symbol string `json:"symbol"`
}
//this process is dedicated to collecting and storing orderbook, raw trade item, and kline data from binance
//differentials are snapshotted every day at midnight
func main(){
	//we need to setup logging, our database client
	util.CheckDir("/tradingbot/logs")
	f,_:= os.OpenFile("/tradingbot/logs/binance_collector_log", os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0777)
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		Password: "",
		DB: 0,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
		})
		
	timescaleClient := pg.Connect(&pg.Options{
        User: "postgres",
		Password: "",
		Addr: "127.0.0.1:5432",
		Database: "tradingbot",
		MaxAge: time.Minute*10,
    })
    defer client.Close()
    defer timescaleClient.Close()
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting data collector")
	//make sure the tables are properly setup
	for _, model := range []interface{}{&binanceOrderBookDiff{},&binanceCandlestick{},&binanceOrderBook{}} {
		err := timescaleClient.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		handleErr(err,"")
	}
	//create timescale hypertables - should be 5 minute chunks now
	_, err := timescaleClient.Exec("SELECT create_hypertable('binance_order_book_diffs','local_time',if_not_exists => true,chunk_time_interval => 60000000000);")
	handleErr(err,"")
	_, err = timescaleClient.Exec("SELECT create_hypertable('binance_candlesticks','local_time',if_not_exists => true,chunk_time_interval => 60000000000);")
	handleErr(err,"")
	_, err = timescaleClient.Exec("SELECT create_hypertable('binance_order_books','local_time',if_not_exists => true,chunk_time_interval => 60000000000);")
	handleErr(err,"")
	
	symbolList, streamDiffString, streamCandleString, symbolSteps := getBinanceSymbols()
	err = client.Set("symbols",symbolList,0).Err()
	handleErr(err, "Setting symbols")
	
	var stats []dayStats
	res, err := http.Get("https://us.binance.com/api/v1/ticker/24hr")
	handleErr(err,"err getting 24 hr stats")
	json.NewDecoder(res.Body).Decode(&stats)
	
	for symbol, value := range symbolSteps {
		err = client.Set(symbol, value,0).Err()
		handleErr(err,"err setting steps")
	}
	
	for _, item := range stats {
		err = client.Set(strings.ToLower(item.Symbol)+":24hr", item.Volume,0).Err()
		handleErr(err,"err setting vol")
	}
	orderbookStringChan := make(chan string)
	candleStringChan := make(chan string)
	go orderbookReader(streamDiffString, client, orderbookStringChan)
	go klineReader(streamCandleString,client, candleStringChan)
	
	//every23 hours, update with any new symbols and redo connections
	for {
		time.Sleep(time.Hour * 23)
		symbolList, streamDiffString, streamCandleString, symbolSteps = getBinanceSymbols()
		res, err := http.Get("https://us.binance.com/api/v1/ticker/24hr")
		handleErr(err,"err getting ticker data")
		json.NewDecoder(res.Body).Decode(&stats)
		for _, item := range stats {
			err = client.Set(strings.ToLower(item.Symbol)+":24hr", item.Volume,0).Err()
			handleErr(err,"err udating 24hr vol")
		}
		//we have to restart all the connections every 24 hours
		for symbol, value := range symbolSteps {
			err = client.Set(symbol, value,0).Err()
			handleErr(err,"err updating symbol steps")
		}
		orderbookStringChan <- streamDiffString
		candleStringChan <- streamCandleString
		ordersnapUpdater(streamDiffString, timescaleClient)
		err = client.Set("symbols",symbolList,0).Err()
		handleErr(err, "Setting symbols")
		client.Publish("updates","symbols")
	}
}
