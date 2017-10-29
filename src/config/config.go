package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// CData is the struct that represent config.json data
type CData struct {
	TelegramToken string `json:"telegram_token"`
	CheckPeriod   int64  `json:"check_period"`
}

// Data is the config data
var Data CData

// ReadConfig is for reading JSON config file and load the data into this package
func ReadConfig() error {
	raw, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("Error when open config err:%v", err)
		return err
	}

	err = json.Unmarshal(raw, &Data)
	if err != nil {
		log.Printf("Error when unmarshal config err:%v", err)
		return err
	}

	return nil
}
