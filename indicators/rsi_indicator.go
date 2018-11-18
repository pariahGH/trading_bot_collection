//every kline interval (1,5,etc), get the most recent indicator value (using local timestamp) for that symbol/interval - if it doesnt exist,
//get the past 10 values and generate (allowing for non existence of past 10 values) then store in db associated with the klines local timestamp
//systems then request from all tables mentioned in config
package main
import (
	"github.com/go-redis/redis"
	"strings"
	"time"
	"../util"
	"strconv"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"log"
	"github.com/markcheno/go-talib"
)

type binanceRsi struct {
	//data item for rsi data
	LocalTime int64
	RSI int
	Period int //for shorter time intervals, we want higher. for longer time intervals, we want shorter
	//key is to emphaise variations of a certain minimum size
}

func analyze(symbol string, dataChan chan rsi){
	//buffer n periods to do calc
}

func main(){
	//setup logging
	util.CheckDir("/tradingbot/logs")
	f,_:= os.OpenFile("/tradingbot/logs/rsi_gen_log", os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0777)
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
	err := timescaleClient.CreateTable(&binanceRsi{}, &orm.CreateTableOptions{
		IfNotExists: true,
	})
	handleErr(err,"")
	_, err := timescaleClient.Exec("SELECT create_hypertable('binance_rsis','local_time',if_not_exists => true,chunk_time_interval => 60000000000);")
	handleErr(err,"")
	//get symbol list
	//launch analyzer for each symbol - probably use the 1m candles to generate the line for resolution
}