package uno

import (
	"errors"
)

type Player struct {
	ID   int
	Hand []Card
}

func (player *Player) CanPlayAny(current Card) bool {
	can := false

	for i := range player.Hand {
		card := player.Hand[i]
		flagMask := card.Flag & current.Flag
		if (current.Flag&Wild) != 0 ||
			(card.Flag&Wild) != 0 ||
			flagMask != 0 {
			can = true
			break
		}
	}

	return can
}

type Deck struct {
	Cards []Card
}

func (deck *Deck) Push(card Card) {
	deck.Cards = append(deck.Cards, card)
}

func (deck *Deck) Peek() Card {
	l := len(deck.Cards)
	if l == 0 {
		return Card{-1, ""}
	}

	res := deck.Cards[l-1]

	return res
}

func (deck *Deck) Pop() (Card, error) {
	l := len(deck.Cards)
	if l == 0 {
		return Card{}, errors.New("empty stack")
	}

	res := deck.Cards[l-1]
	deck.Cards = deck.Cards[:l-1]
	return res, nil
}

type Card struct {
	Flag   int
	CardID string
}

type Direction int
