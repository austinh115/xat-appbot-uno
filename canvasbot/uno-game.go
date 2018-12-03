package canvasbot

import (
	"CanvasBot/config"
	"CanvasBot/global"
	"CanvasBot/uno"
	"CanvasBot/util"
	"encoding/json"
	"fmt"
	"github.com/austinh115/lz-string-go"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type UNOGame struct {
	Players    []*uno.Player
	Deck       uno.Deck
	Pile       uno.Deck
	CurrentPos int
	Direction  uno.Direction
	Started    bool
	bot        *CanvasBot
	Action     string
	ActionById int
	Finished   bool
}

const (
	Clockwise        uno.Direction = 0
	CounterClockwise uno.Direction = 1
)

func NewUNOGame(bot *CanvasBot) *UNOGame {
	return &UNOGame{
		Players:    []*uno.Player{},
		Deck:       uno.Deck{},
		Pile:       uno.Deck{},
		CurrentPos: -1,
		Direction:  Clockwise,
		Started:    false,
		bot:        bot,
		Action:     "",
		Finished:   false,
	}
}

func (game *UNOGame) Start() {
	game.Started = true
	game.NewDeck()

	// Clear game screen and create game container

	resetBoard := map[string]interface{}{
		"name":            "container",
		"destroyChildren": true,
	}

	b, err := json.Marshal(resetBoard)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", resetBoard)
		return
	}

	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})

	// Add container

	container := map[string]interface{}{
		"name": "container",
		"type": "Container",
		"y":    0,
	}

	b, err = json.Marshal(container)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", container)
		return
	}

	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})

	// Game background

	gameBackground := map[string]interface{}{
		"name":   "mygraphics",
		"parent": "container",
		"type":   "Graphics",
		"commands": [][]interface{}{
			{"beginFill", 0x1D8994, 1},
			{"drawRect", 0, 0, 425, 600},
			{"endFill"},
		},
	}

	b, err = json.Marshal(gameBackground)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", gameBackground)
		return
	}

	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})

	// Draw user cards
	// uno.Card width = (130 / 2) = 65,
	// uno.Card height = (182 / 2) = 91

	// Deck of cards on the left

	obj := make([]map[string]interface{}, 0)

	for i := 0; i < 10; i++ {
		cardId := "DeckPile" + strconv.Itoa(i)

		x := float32(55 + (1 * i))

		obj = append(obj, CreateCard(cardId, "CardFront", x, 300, false, 0))
	}

	b, err = json.Marshal(obj)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", b)
	}
	//fmt.Printf("%+v\n", obj)
	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})

	//fmt.Println(game.Players)

	playerInfo := make([]map[string]interface{}, 0)

	playerInfo = append(playerInfo, map[string]interface{}{
		"name": "mystyle",
		"type": "TextStyle",
		"style": map[string]interface{}{
			"fontFamily":      "\"Trebuchet MS\", Helvetica, sans-serif",
			"fontSize":        20,
			"fill":            "#ffffff",
			"stroke":          "#000",
			"strokeThickness": 1,
		},
	})
	playerInfo = append(playerInfo, map[string]interface{}{
		"name": "unostyle",
		"type": "TextStyle",
		"style": map[string]interface{}{
			"fontFamily":      "\"Trebuchet MS\", Helvetica, sans-serif",
			"fontSize":        20,
			"fontWeight":      "bold",
			"fill":            []string{"#D35400", "#C0392B"},
			"stroke":          "#000",
			"strokeThickness": 1,
		},
	})

	// Cards in players hands
	for pNum := range game.Players {
		player := game.Players[pNum]
		game.Draw(player, 7)

		// Add player name and number of cards to top of screen
		playerInfo = append(playerInfo, map[string]interface{}{
			"name":   "playerInfo-" + strconv.Itoa(pNum),
			"type":   "Text",
			"parent": "container",
			"x":      ((425 / 4) * pNum) + (425 / 8),
			"y":      15,
			"text":   game.bot.GetRegnameById(player.ID),
			"style":  "mystyle",
		})
		playerInfo = append(playerInfo, map[string]interface{}{
			"name":   "playerCardCount-" + strconv.Itoa(pNum),
			"type":   "Text",
			"parent": "container",
			"x":      ((425 / 4) * pNum) + (425 / 8),
			"y":      35,
			"text":   7,
			"style":  "mystyle",
		})
	}
	playerInfo = append(playerInfo, map[string]interface{}{
		"name":   "currentTurn",
		"type":   "Text",
		"parent": "container",
		"x":      425 / 2,
		"y":      75,
		"text":   "Current Turn: " + game.bot.GetRegnameById(game.Players[0].ID) + " (" + strconv.Itoa(game.Players[0].ID) + ")",
		"style":  "mystyle",
	})

	b, err = json.Marshal(playerInfo)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", b)
	}
	//fmt.Printf("%+v\n", obj)
	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})

	// Flip first card

	card := game.DrawCardFromDeck()
	if card.Flag == -1 {
		return
	}
	game.Pile.Push(card)
	game.PlayCard(card)

	game.NextTurn()
}

func (game *UNOGame) AddPlayer(id int) bool {
	if game.Started {
		return false
	}
	if len(game.Players) >= 4 {
		return false
	}
	player := uno.Player{ID: id}
	game.Players = append(game.Players, &player)
	return true
}

func (game *UNOGame) RemovePlayer(id int) {
	if game.Started {
		return
	}
	game.ForceRemovePlayer(id)
}

func (game *UNOGame) ForceRemovePlayer(id int) bool {
	for i := range game.Players {
		if game.Players[i].ID == id {
			game.Players = append(game.Players[:i], game.Players[i+1:]...)
			return true
		}
	}
	return false
}

func (game *UNOGame) GetPlayerById(id int) *uno.Player {
	for i := range game.Players {
		if game.Players[i].ID == id {
			return game.Players[i]
		}
	}
	return nil
}

func (game *UNOGame) GetPlayerIndex(id int) int {
	for i := range game.Players {
		if game.Players[i].ID == id {
			return i
		}
	}
	return -1
}

func (game *UNOGame) NewDeck() {
	var deck uno.Deck
	deck.Cards = make([]uno.Card, 0)
	var i uint
	for i = 0; i < 13; i++ {
		numTimes := 2
		if i == 0 {
			numTimes = 1
		}
		for n := 0; n < numTimes; n++ {
			deck.Cards = append(deck.Cards,
				uno.Card{Flag: uno.Blue | (1 << i)},
				uno.Card{Flag: uno.Green | (1 << i)},
				uno.Card{Flag: uno.Yellow | (1 << i)},
				uno.Card{Flag: uno.Red | (1 << i)})
		}
	}
	for i := 0; i < 4; i++ {
		deck.Cards = append(deck.Cards,
			uno.Card{Flag: uno.WildColorChange},
			uno.Card{Flag: uno.WildDraw4})
	}

	deck.Cards = game.Shuffle(deck.Cards)
	game.Deck = deck
}

func (game *UNOGame) Shuffle(cards []uno.Card) []uno.Card {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	shuffled := make([]uno.Card, len(cards))
	n := len(cards)
	for i := 0; i < n; i++ {
		randIndex := r.Intn(len(cards))
		if cards[randIndex].Flag == 0 {
			continue
		}
		shuffled[i] = cards[randIndex]
		cards = append(cards[:randIndex], cards[randIndex+1:]...)
	}
	return shuffled
}

func CreateCard(spriteName string, cardURL string, x float32, y float32, interactive bool, rotation float32) map[string]interface{} {

	f := map[string]interface{}{
		"name":        "card" + spriteName,
		"parent":      "container",
		"type":        "Sprite",
		"imageUrl":    config.Config.UNOCards[cardURL],
		"x":           x,
		"y":           y,
		"scale":       map[string]float32{"x": 0.5, "y": 0.5},
		"interactive": interactive,
		"draggable":   true,
		"buttonMode":  interactive,
		"rotation":    rotation,
	}

	return f
}

func CreateCompressedCard(spriteName string, cardURL string, x float32, y float32, interactive bool, rotation float32) string {
	f := CreateCard(spriteName, cardURL, x, y, interactive, rotation)

	b, err := json.Marshal(f)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", b)
	}

	return LZString.CompressToBase64(string(b))
}

func (game *UNOGame) PlayCard(card uno.Card) {
	// Card pile in the middle
	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"u", game.bot.GetId()},
		{"t", CreateCompressedCard("CenterPile"+card.CardID,
			card.String(),
			212.5,
			300,
			false,
			func() float32 {
				if strings.HasSuffix(card.String(), "6") || strings.HasSuffix(card.String(), "9") {
					return rand.Float32()*0.4 - 0.2
				} else {
					return rand.Float32()
				}
			}())},
	})
}

// Game mechanics

func (game *UNOGame) GetNextTurn() int {
	var nextIndex int
	if game.Direction == Clockwise {
		nextIndex = Mod(game.CurrentPos+1, len(game.Players))
	} else {
		nextIndex = Mod(game.CurrentPos-1, len(game.Players))
	}
	return nextIndex
}

func Mod(d, m int) int {
	var res = d % m
	if (res < 0 && m > 0) || (res > 0 && m < 0) {
		return res + m
	}
	return res
}

func (game *UNOGame) NextTurn() {
	// Change 6 and 9 sprites to be more easily identifiable
doNext:
	action := game.Action
	actionById := game.ActionById
	game.Action = ""
	game.ActionById = 0

	if action == "reverse" {
		game.bot.SendMessage(game.bot.GetRegnameById(actionById) + " reversed the game's direction. It is now going " + (func() string {
			if game.Direction != Clockwise {
				return "clockwise."
			} else {
				return "counter-clockwise."
			}
		})())
		game.ChangeDirection()
	}

	index := game.GetNextTurn()
	game.CurrentPos = index
	//fmt.Println(index, len(game.Players), index%len(game.Players))

	currentPlayer := game.Players[index]

	if action == "skip" {
		game.bot.SendMessage(game.bot.GetRegnameById(actionById) + " skipped " + game.bot.GetRegnameById(currentPlayer.ID) + "'s turn.")
		goto doNext
	}

	//game.bot.SendMessage("It is now " + strconv.Itoa(currentPlayer.ID) + "'s turn.")
	fmt.Println("It is now " + game.bot.GetRegnameById(currentPlayer.ID) + "'s turn.")

	if strings.HasPrefix(action, "draw") {
		numCards, _ := strconv.Atoi(strings.Split(action, "draw")[1])
		game.bot.SendMessage(game.bot.GetRegnameById(actionById) + " forced " + game.bot.GetRegnameById(currentPlayer.ID) + " to draw " + strconv.Itoa(numCards) + " cards.")
		game.Draw(currentPlayer, numCards)
		goto doNext
	}

	// Draw 1 then continue turn
	if !currentPlayer.CanPlayAny(game.Pile.Peek()) {
		//	Player has no cards they can play, draw
		//game.bot.SendMessage(strconv.Itoa(currentPlayer.ID) + " could not play any cards so they were forced to draw.")
		fmt.Println(game.bot.GetRegnameById(currentPlayer.ID) + " could not play any cards so they were forced to draw.")
		game.Draw(currentPlayer, 1)
		if !currentPlayer.CanPlayAny(game.Pile.Peek()) {
			goto doNext
		}
	}

	// Draw until you can play (AIDS)
	//for !currentPlayer.CanPlayAny(game.Pile.Peek()) {
	//	fmt.Println(game.bot.GetRegnameById(currentPlayer.ID) + " could not play any cards so they were forced to draw.")
	//	game.Draw(currentPlayer, 1)
	//}

	game.SendTurnStatus(currentPlayer)
}

func (game *UNOGame) SendTurnStatus(currentPlayer *uno.Player) {
	playerInfo := map[string]interface{}{
		"name": "currentTurn",
		"text": "Current Turn: " + game.bot.GetRegnameById(currentPlayer.ID) + " (" + strconv.Itoa(currentPlayer.ID) + ")",
	}

	b, err := json.Marshal(playerInfo)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", b)
	}
	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})
}

func (game *UNOGame) PlayerTurn(playedCard string, player *uno.Player) {

	cardEvent := strings.SplitN(playedCard, "-", 2)
	if len(cardEvent) != 2 {
		return
	}
	unoCard := cardEvent[1]

	//	Player
	//fmt.Println("Current card on pile:", game.Pile.Peek().String())
	//fmt.Println("Player wants to play:", playedCard)

	hasCard := false
	cardIndex := 0

	//fmt.Println(player.Hand)

	for i := range player.Hand {
		card := player.Hand[i]
		//fmt.Println(card.String(), playedCard)
		if card.CardID == unoCard {
			hasCard = true
			cardIndex = i
			break
		}
	}
	//fmt.Println("Does player have that card?", hasCard)
	if !hasCard {
		// Player doesn't have card
		game.ForceRemovePlayer(player.ID)
		game.bot.SendMessage(game.bot.GetRegnameById(player.ID) + " has been kicked from the game for attempted cheating.")
	} else {
		// Player has the card
		card := player.Hand[cardIndex]
		// Check if can play
		flagMask := card.Flag & game.Pile.Peek().Flag
		fmt.Println(card.CardID, card.Flag, game.Pile.Peek().String(), game.Pile.Peek().Flag, flagMask)
		if (game.Pile.Peek().Flag&uno.Wild) == 0 &&
			(card.Flag&uno.Wild) == 0 &&
			flagMask == 0 {
			// If not a wild card or card doesn't match
			game.bot.SendMessage("You can't play that card.")
			game.ResetCardPosition(player, unoCard)
			return
		}
		if card.Flag&uno.Skip != 0 {
			game.Action = "skip"
		} else if card.Flag&uno.Reverse != 0 {
			game.Action = "reverse"
		} else if card.Flag&uno.Draw2 != 0 {
			game.Action = "draw2"
		} else if card.Flag&uno.Draw4 != 0 {
			game.Action = "draw4"
		}
		if game.Action != "" {
			game.ActionById = player.ID
		}
		if card.Flag&uno.ColorChange != 0 {
			//	Display color change choice
			game.bot.SendMessage("Color changing feature isn't implemented yet (sorry).")
		}
		game.Pile.Push(card)
		player.Hand = append(player.Hand[:cardIndex], player.Hand[cardIndex+1:]...)

		if len(player.Hand) == 1 {
			game.bot.SendMessage(game.bot.GetRegnameById(player.ID) + " has "+strings.ToUpper(global.GAME_NAME)+"!")
			game.DisplayUNO(player.ID, game.CurrentPos)
		}

		if len(player.Hand) == 0 {
			game.bot.SendMessage(game.bot.GetRegnameById(player.ID) + " has won!")
			game.Finished = true
		}

		game.RemovePlayerCard(playedCard, player)
		game.PlayCard(card)
		game.ResetPlayerCards(player)
	}
	if !game.Finished {
		game.NextTurn()
	}
	if game.Finished {
		// Send "You Lose" text to everyone
		// Send "You Win" text to winner to replace "You Lose" text

		//{name:'result', type:'Text',  parent:'game', x:425/2, y:520/2, text:'winner!'}
		resetBoard := []map[string]interface{}{
			{
				"name":            "container",
				"destroyChildren": true,
			},
			{
				"name":   "winner",
				"type":   "Text",
				"parent": "container",
				"x":      425 / 2,
				"y":      520 / 2,
				"text":   game.bot.GetRegnameById(player.ID) + " has won!",
				"style":  "mystyle",
			},
		}

		b, err := json.Marshal(resetBoard)
		if err != nil {
			log.Printf("[message-handler][error] Invalid canvas json: %+v\n", resetBoard)
			return
		}

		game.bot.Send("x", [][]interface{}{
			{"i", "60002"},
			{"u", game.bot.GetId()},
			{"t", LZString.CompressToBase64(string(b))},
		})

		game.bot.UNOGame = nil
	}
}

func (game *UNOGame) DrawCardFromDeck() uno.Card {
	card, err := game.Deck.Pop()
	if err != nil {
		//	Deck is empty
		topCard, _ := game.Pile.Pop()
		game.Deck = game.Pile
		game.Deck.Cards = game.Shuffle(game.Deck.Cards)
		game.Pile = uno.Deck{}
		game.Pile.Push(topCard)

		card, err = game.Deck.Pop()
		if err != nil {
			// Something is fucked up
			game.bot.SendMessage("Unexpected error in restocking the deck.")
			return uno.Card{Flag: -1}
		}
	}

	return card
}

func (game *UNOGame) RemovePlayerCard(s string, player *uno.Player) {
	gameBackground := map[string]interface{}{
		"name":    s,
		"destroy": true,
	}

	b, err := json.Marshal(gameBackground)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", gameBackground)
		return
	}

	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"d", player.ID},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})
}

func (game *UNOGame) ResetCardPosition(player *uno.Player, cardName string) {

	var cardIndex = -1

	for i := range player.Hand {
		if player.Hand[i].CardID == cardName {
			cardIndex = i
			break
		}
	}

	if cardIndex == -1 {
		return
	}

	cardId := "PlayerHand-" + cardName

	x := (425 / (len(player.Hand) + 1)) * (cardIndex + 1)

	gameBackground := map[string]interface{}{
		"name": "card" + cardId,
		"x":    x,
		"y":    532,
	}

	b, err := json.Marshal(gameBackground)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", gameBackground)
		return
	}

	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"d", player.ID},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})
}

func (game *UNOGame) Draw(player *uno.Player, numCards int) {
	for i := 0; i < numCards; i++ {
		topCard := game.DrawCardFromDeck()
		topCard.CardID = util.RandString(10) + "-" + topCard.String()
		player.Hand = append(player.Hand, topCard)
	}

	if len(player.Hand) != 1 {
		removeUno := map[string]interface{}{
			"name":    "displayuno-" + strconv.Itoa(player.ID),
			"destroy": true,
		}

		b, err := json.Marshal(removeUno)
		if err != nil {
			log.Printf("[message-handler][error] Invalid canvas json: %+v\n", removeUno)
			return
		}

		game.bot.Send("x", [][]interface{}{
			{"i", "60002"},
			{"u", game.bot.GetId()},
			{"t", LZString.CompressToBase64(string(b))},
		})
	}

	game.ResetPlayerCards(player)
}

func (game *UNOGame) ResetPlayerCards(player *uno.Player) {
	numCards := len(player.Hand)

	obj := make([]map[string]interface{}, 0)

	for i := range player.Hand {
		card := player.Hand[i]
		//flag := 1<<uint((rand.Intn(4))+13) | (1 << uint(i))
		//card := uno.NewCard(flag).String()

		cardId := "PlayerHand-" + card.CardID
		//fmt.Println(cardId)

		x := float32((425 / (numCards + 1)) * (i + 1))

		obj = append(obj, CreateCard(cardId, card.String(), x, 532, true, 0))
	}
	//fmt.Printf("%+v\n", obj)

	b, err := json.Marshal(obj)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", b)
	}
	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"d", player.ID},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})
	cardCount := map[string]interface{}{
		"name": "playerCardCount-" + strconv.Itoa(game.GetPlayerIndex(player.ID)),
		"text": numCards,
	}

	b, err = json.Marshal(cardCount)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", b)
	}
	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})
}

func (game *UNOGame) ChangeDirection() {
	if game.Direction == Clockwise {
		game.Direction = CounterClockwise
	} else {
		game.Direction = Clockwise
	}
}

func (game *UNOGame) DisplayUNO(id int, index int) {

	displayUno := map[string]interface{}{
		"name":   "displayuno-" + strconv.Itoa(id),
		"type":   "Text",
		"parent": "container",
		"x":      ((425 / 4) * index) + (425 / 8),
		"y":      55,
		"text":   strings.ToUpper(global.GAME_NAME),
		"style":  "unostyle",
	}

	b, err := json.Marshal(displayUno)
	if err != nil {
		log.Printf("[message-handler][error] Invalid canvas json: %+v\n", displayUno)
		return
	}

	game.bot.Send("x", [][]interface{}{
		{"i", "60002"},
		{"u", game.bot.GetId()},
		{"t", LZString.CompressToBase64(string(b))},
	})
}
