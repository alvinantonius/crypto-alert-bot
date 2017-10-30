package main

import (
	"github.com/alvinantonius/crypto-alert-bot/src/config"
	"github.com/alvinantonius/crypto-alert-bot/src/data"
	"github.com/alvinantonius/crypto-alert-bot/src/webhook"
)

func main() {
	config.ReadConfig()
	data.LoadData()

	webhook.Init()
}
