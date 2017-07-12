package main

import (
	"encoding/json"
	"fmt"
	"github.com/bluele/slack"
	"log"
	"net/http"
	"time"
)

var ccApiDomain string = "https://min-api.cryptocompare.com"

var (
	slackToken string = ""
	slackApi          = slack.New(slackToken)
)

var httpClient = http.Client{
	Timeout: time.Second * 10,
}

type coin struct {
	coinType     string
	Eur          float64 `json: "EUR"`
	Usd          float64 `json: "USD"`
	treshold     float64
	pollInterval int
	slackChannel string
}

func getCoinValue(c *coin) {

	res, err := httpClient.Get(fmt.Sprintf("%s/data/price?fsym=%s&tsyms=EUR,USD", ccApiDomain, c.coinType))
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

		if c.Eur <= c.treshold {

			coinValue := fmt.Sprintf("1 %s = %.2fEUR (%.2fUSD)", c.coinType, c.Eur, c.Usd)
			log.Print(coinValue)
			err := slackApi.ChatPostMessage(c.slackChannel, coinValue, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func main() {

	coinCh := make(chan *coin)

	coins := map[string]coin{
		"ETH": coin{
			coinType:     "ETH",
			treshold:     250,
			pollInterval: 500,
			slackChannel: "general",
		},
		"BTC": coin{
			coinType:     "BTC",
			treshold:     2500,
			pollInterval: 200,
			slackChannel: "general",
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
