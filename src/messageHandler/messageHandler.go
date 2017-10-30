package messageHandler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alvinantonius/crypto-alert-bot/src/data"
	"github.com/alvinantonius/crypto-alert-bot/src/messageSender"
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

	switch params[0] {
	case "/start":
		sendHelp(userID)
	case "/alert":
		addWatch(userID, params)
	case "/myalert":
		listWatch(userID)
	case "/remove":
		removeWatch(userID, params)
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

	_, err := data.AddWatch(userID, market, price, when)

	return err
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

	return data.RemoveWatch(watchID)
}

func listWatch(userID int64) error {
	return nil
}
