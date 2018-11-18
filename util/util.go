//utility functions for all agents and strategies
package util

import (
	"time"
	"strconv"
	"os"
	"log"
)

func Timestamp() (stamp string) {
	now := time.Now()
	stamp = strconv.Itoa(now.Year())+"-"+now.Month().String()+"-"+
		strconv.Itoa(now.Day())+"-"+strconv.Itoa(now.Hour())+"-"+
		strconv.Itoa(now.Minute())+"-"+strconv.Itoa(now.Second())
	return 
}

func CheckDir(path string){
	_, e := os.Stat(path)
	if e != nil {
		os.MkdirAll(path,0777)
	}
	return 
}

func handleErr(err error){
	if err != nil{
		log.Println(err)
		panic(err)
	}
}
