package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/bluele/slack"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	cfg        tomlCfg
	slackApi   *slack.Slack
	httpClient = http.Client{
		Timeout: time.Second * 10,
	}
)

type tomlCfg struct {
	CcApiDomain string
	SlackToken  string
	ServerPort  int
	Coins       map[string]coin
}

type coin struct {
	Symbol       string
	Eur          float64 `json: "EUR"`
	Usd          float64 `json: "USD"`
	Treshold     float64
	PollInterval int
	SlackChannel string
}

func getCoinValue(c *coin) {
	res, err := httpClient.Get(fmt.Sprintf("%s/data/price?fsym=%s&tsyms=EUR,USD", cfg.CcApiDomain, c.Symbol))
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
		time.Sleep(time.Duration(c.PollInterval) * time.Second)
	}
}

func sendSlackAlert(coinCh <-chan *coin) {
	for c := range coinCh {
		if c.Eur <= c.Treshold {
			coinValue := fmt.Sprintf("1 %s = %.2fEUR (%.2fUSD)", c.Symbol, c.Eur, c.Usd)
			log.Print(coinValue)
			err := slackApi.ChatPostMessage(c.SlackChannel, coinValue, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi! I have got the following information about your coins:\n\n")
	for symbol, c := range cfg.Coins {
		res, err := httpClient.Get(fmt.Sprintf("%s/data/price?fsym=%s&tsyms=EUR,USD", cfg.CcApiDomain, c.Symbol))
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		output := fmt.Sprintf("%s: %s\n", symbol, string(data))
		fmt.Fprintf(w, output)
	}
}

func main() {

	cfgFile := flag.String("config", "coinmon.toml", "location to TOML config")
	flag.Parse()
	if _, err := toml.DecodeFile(*cfgFile, &cfg); err != nil {
		fmt.Println(err)
		return
	}

	slackApi = slack.New(cfg.SlackToken)
	coinCh := make(chan *coin)

	go sendSlackAlert(coinCh)
	for _, coin := range cfg.Coins {
		go pollCoinValue(coinCh, coin)
	}

	http.HandleFunc("/status", handler)
	http.ListenAndServe(":"+strconv.Itoa(cfg.ServerPort), nil)
}
