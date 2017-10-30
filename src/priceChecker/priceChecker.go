package priceChecker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/alvinantonius/crypto-alert-bot/src/config"
	"github.com/alvinantonius/crypto-alert-bot/src/data"
)

const (
	bitcoinIndonesiaHost = "https://vip.bitcoin.co.id/api"
)

type (
	marketCheck struct {
		URL         string
		Marketplace string
	}

	alerts []data.Watch

	bitcoinIndonesiaResult struct {
		Ticker struct {
			Last string `json:"last"`
		} `json:"ticker"`
	}
)

// stores all option of market that avaialble
var marketCheckList map[string]marketCheck

// store all market that currently on watch
var priceCheckList map[string]bool

// store all alert based on market
// separated between above and below alert
var alertAboveList map[string]alerts
var alertBelowList map[string]alerts

var mutex *sync.Mutex

var httpClient *http.Client

// Make alert sortable
func (a alerts) Len() int {
	return len(a)
}
func (a alerts) Swap(i, j, int) {
	a[i], a[j] = a[j], a[i]
}
func (a alerts) Less(i, j int) bool {
	return a[i].PriceLimit < a[j].PriceLimit
}

// constructing market list and
// initialize all global variables so it will never be nil
func init() {
	marketCheckList = make(map[string]marketCheck)

	marketCheckList["btc-idr"] = marketCheck{
		URL:         fmt.Sprintf("%v/btc_idr/ticker", bitcoinIndonesiaHost),
		Marketplace: "bitcoin.co.id",
	}
	marketCheckList["bch-idr"] = marketCheck{
		URL:         fmt.Sprintf("%v/bch_idr/ticker", bitcoinIndonesiaHost),
		Marketplace: "bitcoin.co.id",
	}
	marketCheckList["xzc-idr"] = marketCheck{
		URL:         fmt.Sprintf("%v/xzc_idr/ticker", bitcoinIndonesiaHost),
		Marketplace: "bitcoin.co.id",
	}

	priceCheckList = make(map[string]bool)

	alertAboveList = make(map[string]alerts)
	alertBelowList = make(map[string]alerts)

	mutex = &sync.Mutex{}

	// init http client for request
	httpClient = &http.Client{Timeout: 1 * time.Second}
}

func (m marketCheckList) CheckPrice() (float64, error) {
	resp, err := httpClient.Get(m.URL)
	if err != nil {
		log.Printf("error do http req for URL:%v err:%v", m.URL, err)
		return 0, err
	}
	defer resp.Body.Close()

	// read data into byte
	body, err := ioutil.ReadAll(resp.Body)

	// prepare variable for unmarshal
	var resData bitcoinIndonesiaResult

	// decode json
	err = json.Unmarshal(body, &resData)
	if err != nil {
		return 0, err
	}

	// convert string to float
	lastPrice, err := strconv.ParseFloat(resData.Ticker.Last, 64)
	if err != nil {
		return 0, err
	}

	return lastPrice, nil
}

// Refresh is for Initialize all market check list
func Refresh() {
	tempPriceCheck := make(map[string]bool)
	tempAlertAbove := make(map[string]alerts)
	tempAlertBelow := make(map[string]alerts)

	for _, user := range data.Get().Users {
		for _, watch := range user.WatchList {

			// init map with certain key if it's never existed before
			if _, ok := tempPriceCheck[watch.Market]; !ok {
				tempPriceCheck[watch.Market] = true
				tempAlertAbove[watch.Market] = []data.Watch{}
				tempAlertBelow[watch.Market] = []data.Watch{}
			}

			if watch.When == "above" {
				tempAlertAbove[watch.Market] = append(tempAlertAbove[watch.Market], watch)
			} else if watch.When == "below" {
				tempAlertBelow[watch.Market] = append(tempAlertBelow[watch.Market], watch)
			}
		}
	}

	// sort all alerts
	for market, alertList := range tempAlertAbove {
		tempAlertAbove[market] = sort.Sort(alertList)
	}
	for market, alertList := range tempAlertBelow {
		tempAlertBelow[market] = sort.Sort(alertList)
	}

	// replace old data with the new one
	mutex.Lock()
	priceCheckList = tempPriceCheck
	mutex.Unlock()

	mutex.Lock()
	alertAboveList = tempAlertAbove
	mutex.Unlock()

	mutex.Lock()
	alertBelowList = tempAlertBelow
	mutex.Unlock()

	return
}

// RunChecker is for run checker for a period of time
// this function is must be only called once
func RunChecker() {

	var wg sync.WaitGroup

	for {

		for market, boolData := range priceCheckList {
			if !boolData {
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				marketCheckList[market].CheckPrice()
			}()
		}

		wg.Wait()

		// wait until check again
		time.Sleep(config.Data.CheckPeriod * time.Second)
	}
}
