package priceChecker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvinantonius/crypto-alert-bot/src/config"
	"github.com/alvinantonius/crypto-alert-bot/src/data"
	"github.com/alvinantonius/crypto-alert-bot/src/messageSender"
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
		err string `json:"error"`
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

// ErrInvalidMarket is error for invalid market
var ErrInvalidMarket = errors.New("invalid market")

// Make alert sortable
func (a alerts) Len() int {
	return len(a)
}
func (a alerts) Swap(i, j int) {
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
	marketCheckList["eth-idr"] = marketCheck{
		URL:         fmt.Sprintf("%v/eth_idr/ticker", bitcoinIndonesiaHost),
		Marketplace: "bitcoin.co.id",
	}
	marketCheckList["etc-idr"] = marketCheck{
		URL:         fmt.Sprintf("%v/etc_idr/ticker", bitcoinIndonesiaHost),
		Marketplace: "bitcoin.co.id",
	}
	marketCheckList["ltc-idr"] = marketCheck{
		URL:         fmt.Sprintf("%v/ltc_idr/ticker", bitcoinIndonesiaHost),
		Marketplace: "bitcoin.co.id",
	}

	priceCheckList = make(map[string]bool)

	alertAboveList = make(map[string]alerts)
	alertBelowList = make(map[string]alerts)

	mutex = &sync.Mutex{}

	// init http client for request
	httpClient = &http.Client{Timeout: 10 * time.Second}
}

func (m marketCheck) CheckPrice() (float64, error) {
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

	if resData.err != "" {
		return 0, ErrInvalidMarket
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
		sort.Sort(alertList)
		tempAlertAbove[market] = alertList
	}
	for market, alertList := range tempAlertBelow {
		sort.Sort(alertList)
		tempAlertBelow[market] = alertList
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
			go marketChecker(market)
			wg.Done()
		}

		wg.Wait()

		// wait until check again
		time.Sleep(time.Duration(config.Data.CheckPeriod) * time.Second)
	}
}

// RegisterChecker is for adding new market watch for user
// alwo this will trigger new market price checker if available
func RegisterChecker(market string) error {
	marketCheckList[market] = marketCheck{
		URL:         fmt.Sprintf("%v/%v/ticker", bitcoinIndonesiaHost, strings.Replace(market, "-", "_", -1)),
		Marketplace: "bitcoin.co.id",
	}

	err := marketChecker(market)
	if err != nil {
		delete(marketCheckList, market)
		return err
	}

	data.SupportedMarket[market] = true
	return nil
}

// marketChecker will run a market checker
func marketChecker(market string) error {
	fmt.Println("do check", market)
	currentPrice, err := marketCheckList[market].CheckPrice()
	if err != nil {
		if err != ErrInvalidMarket {
			log.Printf("fail on check:%v err:%v", market, err)
		}
		return err
	}

	userList := getNotifiedUserList(market, currentPrice)

	if len(userList) > 0 {
		for _, userID := range userList {
			messageSender.NotifyUser(userID, market, currentPrice)
		}

		Refresh()
	}

	return nil
}

// based on current price, check which user is need to be notified
func getNotifiedUserList(market string, currentPrice float64) []int64 {

	// map of userID => watchID
	notifiedUser := make(map[int64]int64)

	// check above list
	if _, ok := alertAboveList[market]; ok {
		for i := 0; i < len(alertAboveList[market]); i++ {
			if alertAboveList[market][i].PriceLimit > currentPrice {
				alertAboveList[market] = alertAboveList[market][i:]
				break
			}

			if alertAboveList[market][i].PriceLimit <= currentPrice {
				userID := alertAboveList[market][i].GetUserID()
				notifiedUser[userID] = alertAboveList[market][i].ID
			}
		}
	}

	// check below list
	if _, ok := alertBelowList[market]; ok {
		for i := len(alertBelowList[market]) - 1; i >= 0; i-- {
			if alertBelowList[market][i].PriceLimit < currentPrice {
				sisa := i
				if i-1 < 0 {
					sisa = 0
				}
				alertBelowList[market] = alertBelowList[market][:sisa]
				break
			}

			if alertBelowList[market][i].PriceLimit >= currentPrice {
				userID := alertBelowList[market][i].GetUserID()
				notifiedUser[userID] = alertBelowList[market][i].ID
			}
		}
	}

	var result []int64

	// compile map to slice of userID
	// also remove watch that has been triggered
	for userID, watchID := range notifiedUser {
		result = append(result, userID)

		data.RemoveWatch(watchID)
	}

	return result
}
