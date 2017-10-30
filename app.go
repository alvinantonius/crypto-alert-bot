package main

import (
	"github.com/alvinantonius/crypto-alert-bot/src/config"
	"github.com/alvinantonius/crypto-alert-bot/src/data"
	"github.com/alvinantonius/crypto-alert-bot/src/priceChecker"
	"github.com/alvinantonius/crypto-alert-bot/src/webhook"
)

func main() {
	config.ReadConfig()
	data.LoadData()

	priceChecker.Refresh()
	go priceChecker.RunChecker()

	webhook.Init()
}
