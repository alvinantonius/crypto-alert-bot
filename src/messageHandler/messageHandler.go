package messageHandler

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/alvinantonius/crypto-alert-bot/src/data"
	"github.com/alvinantonius/crypto-alert-bot/src/messageSender"
	"github.com/alvinantonius/crypto-alert-bot/src/priceChecker"
)

// HandleMessage is for handle command from user
func HandleMessage(userID int64, message string) {

	// trim whitespaces
	message = strings.Trim(message, " ")

	// check if command is valid
	if string(message[0]) != "/" {
		return
	}

	// split params with space
	params := strings.Split(message, " ")

	if len(params) == 0 {
		return
	}

	fmt.Println(params)

	switch params[0] {
	case "/start":
		sendHelp(userID)
	case "/alert":
		addWatch(userID, params)
	case "/myalert":
		listWatch(userID)
	case "/remove":
		removeWatch(userID, params)
	case "/clear":
		clearWatch(userID)
	}

}

func sendHelp(userID int64) error {
	// if user is never registered before, add to user list
	if !data.IsUserRegistered(userID) {
		data.AddUser(userID)
	}

	return nil
}

func addWatch(userID int64, params []string) error {

	if len(params) < 4 {
		return fmt.Errorf("invalid addWatch parameter")
	}

	market := params[1]
	when := params[2]
	priceString := params[3]
	price, _ := strconv.ParseFloat(priceString, 64)

	// check if market is valid

	fmt.Println("add watch")
	_, err := data.AddWatch(userID, market, price, when)

	if err == data.ErrInvalidMarket {
		if err = priceChecker.RegisterChecker(market); err != nil {
			if err != data.ErrInvalidMarket {
				log.Printf("fail register new market %v checker err:%v", market, err)
			}
			messageSender.NotifyInvalidMarket(userID, market)
			return err
		}
	}

	// add watcher using newly registered market
	_, err = data.AddWatch(userID, market, price, when)

	if err == data.ErrInvalidMarket {
		messageSender.NotifyInvalidMarket(userID, market)
		return err
	}

	if err != nil {
		log.Println("fail add watch", err)
		return err
	}

	fmt.Println("refrech price check")
	priceChecker.Refresh()

	messageSender.NotifySuccessAdd(userID, market, price, when)

	return nil
}

func removeWatch(userID int64, params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("invalid removeWatch parameter")
	}

	// validate user
	if !data.IsUserRegistered(userID) {
		return fmt.Errorf("invalid removeWatch userID")
	}

	watchID, _ := strconv.ParseInt(params[1], 10, 64)

	err := data.RemoveWatch(watchID)

	if err != nil {
		log.Println("fail remove watch", err)
		return err
	}

	fmt.Println("refrech price check")
	priceChecker.Refresh()

	messageSender.NotifySuccessRemove(userID)

	return nil
}

func clearWatch(userID int64) error {
	// validate user
	if !data.IsUserRegistered(userID) {
		return fmt.Errorf("invalid clearWatch userID")
	}

	err := data.ClearWatch(userID)
	if err != nil {
		log.Println("fail clear watch :", err)
		return err
	}

	fmt.Println("refresh price check")
	priceChecker.Refresh()

	messageSender.NotifySuccessClear(userID)

	return nil
}

func listWatch(userID int64) error {

	if !data.IsUserRegistered(userID) {
		return nil
	}

	return messageSender.NotifyListAlert(userID)
}
