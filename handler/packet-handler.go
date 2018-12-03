package handler

import (
	"CanvasBot/canvasbot"
	"encoding/json"
	"fmt"
	"github.com/austinh115/lz-string-go"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func HandlePacket(bot *canvasbot.CanvasBot, packet *canvasbot.Packet) {
	switch packet.Tag {
	case "y":
		if bot.LoggingIn {
			bot.Login()
		} else {
			bot.SendJ2(packet)
		}
		break
	case "v":
		bot.JoinData = packet

		bot.Close()

		err := bot.Open()
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

		bot.LoggingIn = false
		bot.JoinChat()
		break
	case "m":
		if bot.Done {
			MessageHandler(bot, packet)
		}
		break
	case "u":
		if packet.HasAttribute("N") {
			id, err := strconv.Atoi(packet.GetAttribute("u"))
			if err != nil {
				fmt.Println("Error with id:", id)
			}
			bot.Players[strings.ToLower(packet.GetAttribute("N"))] = id
		}
		break
	case "l":
		pId, err := strconv.Atoi(packet.GetAttribute("u"))
		if err != nil {
			fmt.Println("Error with id:", pId)
			break
		}
		for name, id := range bot.Players {
			if id == pId {
				delete(bot.Players, name)
				break
			}
		}
		break
	case "g":
		// user joins app
		// <g u="638877683" x="60002" />
		break
	case "done":
		bot.Done = true
	case "x":
		data := packet.GetAttribute("t")
		if data == "" {
			return
		}
		data, err := LZString.DecompressFromBase64(data)
		if err != nil {
			log.Println("[packet-handler][canvas] Invalid response. Error:", err)
			return
		}
		var cmd interface{}
		err = json.Unmarshal([]byte(data), &cmd)
		if err != nil {
			log.Println("[packet-handler][canvas] Invalid response. Error:", err)
			return
		}
		fmt.Printf("%+v\n", cmd)

		event := cmd.(map[string]interface{})

		if event["action"] == "dragEnd" {
			if bot.UNOGame.Started != true {
				return
			}
			player := packet.GetAttribute("u")
			id, _ := strconv.Atoi(player)
			unoPlayer := bot.UNOGame.GetPlayerById(id)
			if unoPlayer != nil {
				playerIndex := bot.UNOGame.GetPlayerIndex(id)
				cardName := strings.SplitN(event["name"].(string), "-", 2)[1]
				if bot.UNOGame.CurrentPos != playerIndex {
					//	Not their turn, ignore
					fmt.Println("Current turn:", bot.UNOGame.Players[bot.UNOGame.CurrentPos].ID)
					bot.UNOGame.ResetCardPosition(unoPlayer, cardName)
					return
				}
				if event["x"].(float64) > 195 &&
					event["x"].(float64) < 295 &&
					event["y"].(float64) < 325 &&
					event["y"].(float64) > 280 {
					// They've dragged onto the center pile
					// Check if valid card
					bot.UNOGame.PlayerTurn(event["name"].(string), unoPlayer)
				} else {
					fmt.Println("Card not in range")
					if err == nil {
						bot.UNOGame.ResetCardPosition(unoPlayer, cardName)
					}
				}
			} else {
				fmt.Println("Invalid player:", id)
			}
		}
		break
	}

}
