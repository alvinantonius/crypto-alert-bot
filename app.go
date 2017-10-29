package main

import (
	"github.com/alvinantonius/pantau-koin-bot/src/config"
	"github.com/alvinantonius/pantau-koin-bot/src/data"
)

func main() {
	config.ReadConfig()
	data.LoadData()
}
