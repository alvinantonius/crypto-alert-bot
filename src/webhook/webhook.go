package webhook

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alvinantonius/crypto-alert-bot/src/messageHandler"

	"github.com/julienschmidt/httprouter"
)

type (
	// Data is representing updates data from telegram
	Data struct {
		ID      int64   `json:"update_id"`
		Message Message `json:"message"`
	}

	// Message is representing message data from telegram
	Message struct {
		ID   int64  `json:"message_id"`
		From User   `json:"from"`
		Text string `json:"text"`
	}

	// User is representing user data from telegram
	User struct {
		ID    int64 `json:"id"`
		IsBot bool  `json:"is_bot"`
	}
)

// Handler is the http handler that will get all messages and updates from telegram
func Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// prepare data obj
	var data Data

	// decode JSON input
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&data)

	// return if bot
	if data.Message.From.IsBot {
		return
	}
}

// Init is for initializing webhook
func Init() {
	router := httprouter.New()

	router.POST("/crypto-alert/v1/updates", Handler)

	log.Fatal(http.ListenAndServe(":8000", router))
}
