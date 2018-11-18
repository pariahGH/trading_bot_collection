//contains types and functions for the binance collector
package main
import (
	"log"
	"golang.org/x/net/websocket"
	"strings"
	"encoding/json"
	"net/http"
	"github.com/go-redis/redis"
	"github.com/go-pg/pg"
	"time"
)

func handleErr(err error, what string){
	if err != nil{
		log.Println(what)
		log.Println(err)
		panic(err)
	}
}

var intervals = "1m 1h 1d"
func getBinanceSymbols()(symbolList, streamDiffString, streamCandleString string, symbolSteps map[string]string){
	var eInfo exchangeInfo
	res, err := http.Get("https://api.binance.com/api/v1/exchangeInfo")
	symbolList = ""
	streamDiffString = ""
	streamCandleString = ""
	symbolSteps = make(map[string]string)
	handleErr(err,"getting exchange info")
	json.NewDecoder(res.Body).Decode(&eInfo)
	for i := 0; i< len(eInfo.Symbols); i++{
		//for every symbol, find the LOT_SIZE filter and add it to our map
		if eInfo.Symbols[i].Status == "TRADING" && strings.HasSuffix(eInfo.Symbols[i].Symbol,"ETH"){
			symbol := eInfo.Symbols[i].Symbol
			symbolList+=strings.ToLower(symbol)+","
			streamDiffString+=strings.ToLower(symbol)+"@depth/"
			for _, interval := range strings.Split(intervals," "){
				streamCandleString+=strings.ToLower(symbol)+"@kline_"+interval+"/"
			}
			for j := 0; j<len(eInfo.Symbols[i].Filters); j++{
				//we only want ones that are trading and are against ETH, and we want the lotsize
				if eInfo.Symbols[i].Filters[j].FilterType == "LOT_SIZE" {
					symbolSteps[strings.ToLower(symbol)+":step"] = eInfo.Symbols[i].Filters[j].StepSize
				}
				break
			}
		}
	}
	return
}

func orderbookReader(streamString string, client *redis.Client, streamStringChan chan string){
	conn, err := websocket.Dial("wss://stream.binance.com:9443/ws/"+streamString, "", "http://localhost")
	handleErr(err,"dialing orderbook stream")
	var data binanceOrderBookDiff
	
	start := time.Now().UnixNano()
	buffer := make([]binanceOrderBookDiff,0)
	
	for{
		select{
			case c := <- streamStringChan:
				log.Println("opened new orderbook stream connection")
				conn.Close()
				conn, err = websocket.Dial("wss://stream.binance.com:9443/ws/"+c, "", "http://localhost")
				handleErr(err,"dialing new orderbook stream")
			default:
				websocket.JSON.Receive(conn, &data)
				//need to update the appropiate snap
				data.LocalTime = time.Now().UTC().UnixNano()
				dataString,err := json.Marshal(data)
				handleErr(err,"marshaling data for orderbook")
				err = client.Publish("binance.orderbookdiff."+strings.ToLower(data.Symbol),string(dataString)).Err()
				handleErr(err,"publishing orderbook")
				buffer = append(buffer, data)
				if (time.Now().UnixNano() - start) > int64(time.Minute * 10){
					timescaleClient := pg.Connect(&pg.Options{
						User: "postgres",
						Password: "",
						Addr: "127.0.0.1:5432",
						Database: "tradingbot",
					})
					err = timescaleClient.Insert(&buffer)
					handleErr(err,"inserting orderbook")
					err = timescaleClient.Close()
					handleErr(err,"closing conn, orderbook")
					start = time.Now().UnixNano()
					buffer = make([]binanceOrderBookDiff,0)
				}
		}
	}
}

func ordersnapUpdater(streamString string, timescaleClient *pg.DB){
	for _, symbol := range strings.Split(streamString, "@depth/"){
		//get the snapshot
		var data binanceOrderBook
		resp, err := http.Get("https://us.binance.com/api/v1/depth?symbol="+symbol+"&limit=1000")
		handleErr(err,"err getting orderbook snapshots")
		json.NewDecoder(resp.Body).Decode(&data)
		data.LocalTime = time.Now().UTC().UnixNano()
		//save to db
		err = timescaleClient.Insert(&data)
		handleErr(err,"err saving orderbook snapshot to db")
		log.Println("saved snapshot for "+symbol)
		time.Sleep(time.Second)
	}
	return
}

func klineReader(streamString string, client *redis.Client,streamStringChan chan string){
	conn, err := websocket.Dial("wss://stream.binance.com:9443/ws/"+streamString, "", "http://localhost")
	handleErr(err,"err dialing kline stream")
	var data binanceCandlestick
	
	start := time.Now().UnixNano()
	buffer := make([]binanceCandlestick,0)
	
	for{
		select{
			case c := <- streamStringChan:
				log.Println("opened new candle stream connection")
				err = conn.Close()
				conn, err = websocket.Dial("wss://stream.binance.com:9443/ws/"+c, "", "http://localhost")
				handleErr(err,"err dialing kline stream")
			default:
				websocket.JSON.Receive(conn, &data)
				data.LocalTime = time.Now().UTC().UnixNano()
				dataString,err := json.Marshal(data)
				handleErr(err,"err marshaling kline")
				if data.Kline.Closed { //only push closed ones! anything else is something thathasnt finished yet and shouldnt be used
					err = client.Publish("binance.kline."+strings.ToLower(data.Symbol)+"."+data.Kline.Interval,string(dataString)).Err()
					handleErr(err,"error publishing kline")
					buffer = append(buffer, data)
					if (time.Now().UnixNano() - start) > int64(time.Minute * 10){
						timescaleClient := pg.Connect(&pg.Options{
							User: "postgres",
							Password: "",
							Addr: "127.0.0.1:5432",
							Database: "tradingbot",
						})
						err = timescaleClient.Insert(&buffer)
						handleErr(err,"inserting kline")
						err = timescaleClient.Close()
						handleErr(err,"closing conn, kline")
						start = time.Now().UnixNano()
						buffer = make([]binanceCandlestick,0)
					}
				}
		}
	}
}

