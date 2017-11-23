package messageSender

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alvinantonius/crypto-alert-bot/src/config"
	"github.com/alvinantonius/crypto-alert-bot/src/data"
)

const (
	telegramHost = "https://api.telegram.org/bot"
)

var httpClient *http.Client

func init() {
	// init http client for request
	httpClient = &http.Client{Timeout: 10 * time.Second}
}

// NotifyUser is for send message to used about current market price
func NotifyUser(userID int64, market string, price float64) error {

	market = strings.ToUpper(market)
	priceString := strconv.FormatFloat(price, 'f', -1, 64)

	message := fmt.Sprintf("%v = %v", market, priceString)

	return sendMessage(userID, message)
}

// NotifyInvalidMarket is for notifying user that he/she input wrong market code
func NotifyInvalidMarket(userID int64, market string) error {
	market = strings.ToUpper(market)

	message := fmt.Sprintf("Invalid '%v' market code\n\n List of available market is:\n", market)
	for market := range data.SupportedMarket {
		message = fmt.Sprintf("%v\n%v", message, market)
	}

	return sendMessage(userID, message)
}

// NotifyListAlert is for get current list of alert
func NotifyListAlert(userID int64) error {

	user, err := data.GetUser(userID)
	if err != nil {
		log.Println("fail get user")
		return err
	}

	var message string
	if len(user.WatchList) == 0 {
		message = "you have no market price alert right now"
	} else {
		message = "Here is your market price alert list:\n"
		for _, watch := range user.WatchList {
			alertText := fmt.Sprintf("%v %v %v | id=%v", strings.ToUpper(watch.Market), watch.When, strconv.FormatFloat(watch.PriceLimit, 'f', -1, 64), watch.ID)
			message = fmt.Sprintf("%v\n%v", message, alertText)
		}
	}

	return sendMessage(userID, message)
}

// NotifySuccessAdd is for callback that user alert is recorded
func NotifySuccessAdd(userID int64, market string, price float64, when string) error {
	market = strings.ToUpper(market)
	priceString := strconv.FormatFloat(price, 'f', -1, 64)

	message := fmt.Sprintf("Created new alert for:\n%v %v %v", market, when, priceString)

	return sendMessage(userID, message)
}

// NotifySuccessRemove is for callback that remove is success
func NotifySuccessRemove(userID int64) error {
	return sendMessage(userID, "Alert removed")
}

// NotifySuccessClear is for callback that clear is success
func NotifySuccessClear(userID int64) error {
	return sendMessage(userID, "All alert cleared")
}

func sendMessage(userID int64, message string) error {

	messageEnc := url.Values{}
	messageEnc.Add("text", message)
	message = messageEnc.Encode()

	requrl := fmt.Sprintf("%v%v/sendMessage?chat_id=%v&%v", telegramHost, config.Data.TelegramToken, userID, message)

	resp, err := httpClient.Get(requrl)
	if err != nil {
		log.Printf("Fail send message to userID:%v err:%v", userID, err)
		return err
	}
	defer resp.Body.Close()

	return nil
}
