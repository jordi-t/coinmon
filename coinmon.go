package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	apiDomain string = "https://min-api.cryptocompare.com"
)

type coin struct {
	coinType     string
	Eur          float64 `json: "EUR"`
	Usd          float64 `json: "USD"`
	treshold     float64
	pollInterval int
}

var httpClient = http.Client{
	Timeout: time.Second * 10,
}

func getCoinValue(c *coin) {

	res, err := httpClient.Get(fmt.Sprintf("%s/data/price?fsym=%s&tsyms=EUR,USD", apiDomain, c.coinType))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	d := json.NewDecoder(res.Body)
	err = d.Decode(&c)
	if err != nil {
		log.Fatal(err)
	}
}

func pollCoinValue(coinCh chan<- *coin, c coin) {

	for {
		getCoinValue(&c)
		coinCh <- &c
		time.Sleep(time.Duration(c.pollInterval) * time.Second)
	}
}

func sendSlackAlert(coinCh <-chan *coin) {

	for {
		c := <-coinCh
		log.Print(fmt.Sprintf("1 %s = %.2fEUR (%.2fUSD)", c.coinType, c.Eur, c.Usd))
	}
}

func main() {

	coinCh := make(chan *coin)

	coins := map[string]coin{
		"ETH": coin{
			coinType:     "ETH",
			treshold:     250,
			pollInterval: 5,
		},
		"BTC": coin{
			coinType:     "BTC",
			treshold:     2500,
			pollInterval: 2,
		},
	}

	go sendSlackAlert(coinCh)

	for _, coin := range coins {
		go pollCoinValue(coinCh, coin)
	}

	fmt.Println("\nDuring execution, press any key to close\n")
	var input string
	fmt.Scanln(&input)
}
