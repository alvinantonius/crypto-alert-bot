package main

import (
	"github.com/alvinantonius/crypto-alert-bot/src/config"
	"github.com/alvinantonius/crypto-alert-bot/src/data"
)

func main() {
	config.ReadConfig()
	data.LoadData()
}
