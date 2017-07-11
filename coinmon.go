package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	apiDomain        string  = "https://min-api.cryptocompare.com"
	ethAlertTreshold float64 = 200
	btcAlertTreshold float64 = 3000
)

type coinValue struct {
	coinType string
	Eur      float64 `json: "EUR"`
	Usd      float64 `json: "USD"`
}

func getCoinValue(coinCh chan<- *coinValue, c coinValue) {

	coinClient := http.Client{
		Timeout: time.Second * 10,
	}

	for {

		res, getErr := coinClient.Get(fmt.Sprintf("%s/data/price?fsym=%s&tsyms=EUR,USD", apiDomain, c.coinType))
		if getErr != nil {
			log.Fatal(getErr)
		}

		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}

		// ??? use defer here ???
		defer res.Body.Close()

		//coinValue := coinValue{}

		jsonErr := json.Unmarshal(body, &c)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}

		coinCh <- &c

		time.Sleep(5 * time.Second)

	}
}

func sendSlackAlert(ethCh <-chan *coinValue, btcCh <-chan *coinValue) {

	for {
		select {

		case ethValue := <-ethCh:
			log.Print("@ETH chan")
			if ethValue.Eur <= ethAlertTreshold {
				log.Print(fmt.Sprintf("1 ETH = %.2fEUR (%.2fUSD)", ethValue.Eur, ethValue.Usd))
			}

		case btcValue := <-btcCh:
			log.Print("@BTC chan")
			if btcValue.Eur <= btcAlertTreshold {
				log.Print(fmt.Sprintf("1 BTC = %.2fEUR (%.2fUSD)", btcValue.Eur, btcValue.Usd))
			}
		}
	}
}

func main() {

	ethCh := make(chan *coinValue)
	btcCh := make(chan *coinValue)

	go sendSlackAlert(ethCh, btcCh)
	go getCoinValue(ethCh, coinValue{coinType: "ETH"})
	go getCoinValue(btcCh, coinValue{coinType: "BTC"})

	fmt.Println("\nDuring execution, press any key to close\n")
	var input string
	fmt.Scanln(&input)
}
