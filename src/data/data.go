package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"
)

type (
	// Container is representing main data object
	Container struct {
		Users         []User `json:"users"`
		WatchSequence int64  `json:"watch_sequence"`
	}

	// User is representing user data
	User struct {
		ID        int64   `json:"id"`
		WatchList []Watch `json:"watch_list"`
	}

	// Watch is representing user watch data
	Watch struct {
		ID         int64   `json:"id"`
		Market     string  `json:"market"`
		PriceLimit float64 `json:"price_limit"`
		When       string  `json:"notify_when"`
	}
)

var mutex *sync.Mutex

var data Container

// contains user index relative to data(Container)
// map[user_id]=>index of this user in data(container)
var userIndex map[int64]int

// contains watch index relative to data(container)
// map[watch_id]=>[user-index]-[watch-index in user]
var watchIndex map[int64]string

// LoadData is for loading data from data.json file into this package
func LoadData() error {
	raw, err := ioutil.ReadFile("data.json")
	if err != nil {
		log.Printf("Error when open data err:%v", err)
		return err
	}

	err = json.Unmarshal(raw, &data)
	if err != nil {
		log.Printf("Error when unmarshal data err:%v", err)
		return err
	}

	mutex = &sync.Mutex{}

	indexData()

	return nil
}

func indexData() {
	// make sure map is never nil
	userIndex = make(map[int64]int)
	watchIndex = make(map[int64]string)

	for uIndex, user := range data.Users {
		userIndex[user.ID] = uIndex
		for wIndex, watch := range user.WatchList {
			watchIndex[watch.ID] = fmt.Sprintf("%v-%v", uIndex, wIndex)
		}
	}
}

// SaveData is to store current data state into data.json
func SaveData() error {
	dataByte, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Println("Error fail unmarshal data ->", err)
		return err
	}

	err = ioutil.WriteFile("data.json", dataByte, 0644)
	if err != nil {
		log.Println("Error write to data.json file:", err)
		return err
	}

	return nil
}

// Get is for get data
func Get() Container {
	return data
}

// IsUserRegistered is for check whether this user_id is already recorded or not
func IsUserRegistered(userID int64) bool {
	_, ok := userIndex[userID]
	return ok
}

// AddUser is for adding new user
func AddUser(userID int64) {

	if IsUserRegistered(userID) {
		return
	}

	uIndex := len(data.Users)

	mutex.Lock()
	defer mutex.Unlock()

	data.Users = append(data.Users, User{ID: userID})
	userIndex[userID] = uIndex

	SaveData()

	return
}

// AddWatch is for adding new market watch for user
func AddWatch(userID int64, market string, price float64, when string) (Watch, error) {
	if !IsUserRegistered(userID) {
		return Watch{}, fmt.Errorf("userID not found")
	}

	// get user index index
	uIndex := userIndex[userID]

	// get new watch index
	wIndex := len(data.Users[uIndex].WatchList)

	market = strings.ToLower(market)
	// TODO validate market string

	// validate when
	when = strings.ToLower(when)
	if when != "above" && when != "below" {
		return Watch{}, fmt.Errorf("invalid price watch threshold")
	}

	// validate price > 0
	if price <= 0 {
		return Watch{}, fmt.Errorf("invalid price")
	}

	// check if this watch is already exist
	for _, watch := range data.Users[uIndex].WatchList {
		if watch.Market == market &&
			watch.PriceLimit == price &&
			watch.When == when {
			return Watch{}, fmt.Errorf("Duplicate watch")
		}
	}

	mutex.Lock()
	data.WatchSequence++
	watchID := data.WatchSequence
	mutex.Unlock()

	// add to user watch
	watchData := Watch{
		ID:         watchID,
		Market:     market,
		PriceLimit: price,
		When:       when,
	}

	mutex.Lock()
	defer mutex.Unlock()
	data.Users[uIndex].WatchList = append(data.Users[uIndex].WatchList, watchData)

	// add to watch index
	watchIndex[watchID] = fmt.Sprintf("%v-%v", uIndex, wIndex)

	SaveData()

	return watchData, nil
}

// RemoveWatch is for removing watch
func RemoveWatch(watchID int64) error {
	// validate watchID
	if _, ok := watchIndex[watchID]; !ok {
		return fmt.Errorf("invalid watch id")
	}

	// get user and watch index
	userWatchIndex := watchIndex[watchID]
	indexes := strings.Split(userWatchIndex, "-")

	// parse all index to integer
	uIndex, _ := strconv.Atoi(indexes[0])
	wIndex, _ := strconv.Atoi(indexes[1])

	// remove watch
	mutex.Lock()
	defer mutex.Unlock()
	data.Users[uIndex].WatchList = append(
		data.Users[uIndex].WatchList[:wIndex],
		data.Users[uIndex].WatchList[wIndex+1:]...,
	)

	SaveData()

	return nil
}

// ListWatch is for retrieving certain user alert list
func ListWatch(userID int64) ([]Watch, error) {
	if !IsUserRegistered(userID) {
		return []Watch{}, fmt.Errorf("userID not exist")
	}

	uIndex := userIndex[userID]

	return data.Users[uIndex].WatchList, nil
}

// GetUserID is for get this watch userID
func (w Watch) GetUserID() int64 {
	userWatchIndex := watchIndex[w.ID]
	indexes := strings.Split(userWatchIndex, "-")

	// parse all index to integer
	uIndex, _ := strconv.Atoi(indexes[0])

	return data.Users[uIndex].ID
}
