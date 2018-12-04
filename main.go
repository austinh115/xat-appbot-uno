package main

import (
	"./canvasbot"
	"./config"
	"./handler"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	config.Load()
	rand.Seed(time.Now().Unix())
}

func main() {
	var bot = connect()

	bot.JoinChat()
	// Wait for a CTRL-C
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	bot.Close()
}

func connect() *canvasbot.CanvasBot {
	var bot, err = canvasbot.New(&config.Config, handler.HandlePacket)
	log.Println("[main][connect] Connecting to xat...")
	err = bot.Open()
	attempts := 3
	for err != nil && attempts > 0 {
		log.Printf("[main][connect] Failed to establish connection with xat: %s\n", err)
		log.Println("[main][connect] Retrying in 3 seconds")
		time.Sleep(3 * time.Second)
		err = bot.Open()
		attempts--
	}
	if attempts == 0 {
		log.Fatalln("[main][connect] Unable to establish connection")
		os.Exit(1)
	}
	log.Println("[main][connect] Connected to xat successfully")
	return bot
}
