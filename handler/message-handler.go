package handler

import (
	"CanvasBot/canvasbot"
	"CanvasBot/global"
	"CanvasBot/uno"
	"log"
	"strconv"
	"strings"
)

type CommandInfo struct {
	Execute func(bot *canvasbot.CanvasBot, message *canvasbot.Packet, cmd string, query string, arguments []string) bool
}

var commands = map[string]CommandInfo{
	"say": {
		Execute: func(bot *canvasbot.CanvasBot, message *canvasbot.Packet, cmd string, query string, arguments []string) bool {
			bot.SendMessage(strings.Join(arguments[1:], " "))
			return false
		},
	},
	"app": {
		Execute: func(bot *canvasbot.CanvasBot, message *canvasbot.Packet, cmd string, query string, arguments []string) bool {
			// <x i="30008" u="638877683" t="j" />
			bot.Send("x", [][]interface{}{
				{"i", arguments[1]},
				{"u", bot.GetId()},
				{"t", ""},
			})
			return false
		},
	},
	global.GAME_NAME: {
		Execute: func(bot *canvasbot.CanvasBot, message *canvasbot.Packet, cmd string, query string, arguments []string) bool {
			// <x i="30008" u="638877683" t="j" />

			//f := map[string]interface{}{
			//	"name": "background",
			//	"type": "Sprite",
			//	"imageUrl": "https://i.imgur.com/mQa7P4D.jpg",
			//	"anchor": map[string]int{"x": 0, "y": 0 },
			//	"width": 425,
			//	"height": 600,
			//}

			switch arguments[1] {
			case "removeuser":
				if bot.UNOGame == nil || !bot.UNOGame.Started {
					bot.SendMessage("A game of "+global.GAME_NAME+" is not started.")
					return false
				}
				id, err := strconv.Atoi(arguments[2])
				if err != nil {
					id = bot.GetPlayerByRegname(arguments[2])
					if id != -1 {
						goto kickUser
					}
					bot.SendMessage("Invalid user. Error: " + err.Error())
					return false
				}
			kickUser:
				hand := bot.UNOGame.GetPlayerById(id).Hand
				inGame := bot.UNOGame.ForceRemovePlayer(id)
				if inGame {
					bot.SendMessage(bot.GetRegnameById(id) + "(" + strconv.Itoa(id) + ") has left. Their cards have been returned to the deck and the deck has been shuffled.")
					bot.UNOGame.Deck.Cards = append(bot.UNOGame.Deck.Cards, hand...)
					bot.UNOGame.Deck.Cards = bot.UNOGame.Shuffle(bot.UNOGame.Deck.Cards)
				} else {
					bot.SendMessage("Player not found.")
				}
				break
			case "start":
				if bot.UNOGame != nil && bot.UNOGame.Started {
					bot.SendMessage("A game of "+global.GAME_NAME+" has already started.")
					return false
				}
				bot.UNOGame = canvasbot.NewUNOGame(bot)

				players := arguments[2:]

				for i := range players {
					id, err := strconv.Atoi(players[i])
					if err != nil {
						id = bot.GetPlayerByRegname(players[i])
						if id != -1 {
							goto addPlayer
						}
						bot.SendMessage("Invalid user " + players[i] + ". Error, read console.")
						log.Printf("[message-handler][error] %s\n", err.Error())
						continue
					}
				addPlayer:
					ok := bot.UNOGame.AddPlayer(id)
					if ok {
						bot.SendMessage(players[i] + " has been added to the game.")
					} else {
						bot.SendMessage(players[i] + " was not added to the game because it is full (Maximum 4 players).")
					}
				}

				if len(bot.UNOGame.Players) == 0 {
					bot.SendMessage("You need at least 1 player to play...")
					return false
				}

				bot.UNOGame.Start()
				break
			case "restart":
				if bot.UNOGame == nil || !bot.UNOGame.Started {
					bot.SendMessage("A game of "+global.GAME_NAME+" is not started.")
					return false
				}
				var players []*uno.Player
				if bot.UNOGame != nil {
					players = bot.UNOGame.Players
					for i := range players {
						players[i].Hand = make([]uno.Card, 0)
					}
				}
				bot.UNOGame = nil
				bot.UNOGame = canvasbot.NewUNOGame(bot)
				bot.UNOGame.Players = players
				bot.SendMessage("The current game of "+global.GAME_NAME+" has been restarted.")
				bot.UNOGame.Start()
				break
			case "stop":
				if bot.UNOGame == nil || !bot.UNOGame.Started {
					bot.SendMessage("A game of "+global.GAME_NAME+" is not started.")
					return false
				}
				bot.UNOGame = nil
				bot.SendMessage("The current game of "+global.GAME_NAME+" has been stopped.")
				break
			case "status":
				if bot.UNOGame == nil || !bot.UNOGame.Started {
					bot.SendMessage("A game of "+global.GAME_NAME+" is not started.")
					return false
				}
				bot.SendMessage("Current turn: " + strconv.Itoa(bot.UNOGame.Players[bot.UNOGame.CurrentPos].ID) + " | Current card: " + bot.UNOGame.Pile.Peek().String())
				break
			}

			return false
		},
	},
}

func MessageHandler(bot *canvasbot.CanvasBot, message *canvasbot.Packet) {
	if !strings.HasPrefix(message.GetAttribute("u"), global.OWNER_ID) {
		return
	}

	text := message.GetAttribute("t")
	if strings.HasPrefix(text, global.COMMAND_PREFIX) {
		arguments := strings.Split(strings.Trim(strings.Replace(text, global.COMMAND_PREFIX, "", 1), " "), " ")
		query := strings.Replace(text, global.COMMAND_PREFIX+arguments[0]+" ", "", 1)
		cmd := strings.ToLower(arguments[0])

		commandInfo, isKeyPresent := commands[cmd]
		if isKeyPresent {
			commandInfo.Execute(bot, message, cmd, query, arguments)
		}
	}
}
