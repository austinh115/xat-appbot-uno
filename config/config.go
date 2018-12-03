package config

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	Chat    int `json:"Chat"`
	BotInfo struct {
		Username string `json:"Username"`
		ApiKey   string `json:"ApiKey"`
		Name     string `json:"Name"`
		Avatar   string `json:"Avatar"`
		Home     string `json:"Home"`
	} `json:"BotInfo"`
	UNOCards map[string]string `json:"UNOCards"`
}

var Config Configuration

func Load() {
	log.Println("[config][Load] Loading configuration file")
	configFile, _ := os.Open("config.json")
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatalln("[config][init] Unable to decode configuration:", err)
	}
	Config = configuration
}

func Save() {
	configFile, err := os.Create("config.json")
	defer configFile.Close()
	if err != nil {
		log.Fatalln("[config][Save] Unable to open configuration file:", err)
	}
	configFile.WriteString(string(ToJson()))
}

func ToJson() string {
	data, _ := json.MarshalIndent(Config, "", "  ")
	return string(data)
}
