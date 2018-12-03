package uno

import (
	"fmt"
)

const (
	Zero    int = 1 << 0
	One     int = 1 << 1
	Two     int = 1 << 2
	Three   int = 1 << 3
	Four    int = 1 << 4
	Five    int = 1 << 5
	Six     int = 1 << 6
	Seven   int = 1 << 7
	Eight   int = 1 << 8
	Nine    int = 1 << 9
	Draw2   int = 1 << 10
	Reverse int = 1 << 11
	Skip    int = 1 << 12

	Blue   int = 1 << 13
	Green  int = 1 << 14
	Red    int = 1 << 15
	Yellow int = 1 << 16
	Wild   int = 1 << 17

	ColorChange int = 1 << 18
	Draw4       int = 1 << 19

	WildColorChange = Wild | ColorChange
	WildDraw4       = Wild | ColorChange | Draw4
)

func (card Card) String() string {
	out := ""
	if card.Flag&Blue != 0 {
		out += "Blue"
	} else if card.Flag&Green != 0 {
		out += "Green"
	} else if card.Flag&Yellow != 0 {
		out += "Yellow"
	} else if card.Flag&Red != 0 {
		out += "Red"
	} else if card.Flag&Wild != 0 {
		out += "Wild"
	}

	if card.Flag&Zero != 0 {
		//out += "Zero"
		out += "0"
	} else if card.Flag&One != 0 {
		//out += "One"
		out += "1"
	} else if card.Flag&Two != 0 {
		//out += "Two"
		out += "2"
	} else if card.Flag&Three != 0 {
		//out += "Three"
		out += "3"
	} else if card.Flag&Four != 0 {
		//out += "Four"
		out += "4"
	} else if card.Flag&Five != 0 {
		//out += "Five"
		out += "5"
	} else if card.Flag&Six != 0 {
		//out += "Six"
		out += "6"
	} else if card.Flag&Seven != 0 {
		//out += "Seven"
		out += "7"
	} else if card.Flag&Eight != 0 {
		//out += "Eight"
		out += "8"
	} else if card.Flag&Nine != 0 {
		//out += "Nine"
		out += "9"
	} else if card.Flag&Draw2 != 0 {
		out += "Draw2"
	} else if card.Flag&Reverse != 0 {
		out += "Reverse"
	} else if card.Flag&Skip != 0 {
		out += "Skip"
	}

	if card.Flag&Draw4 != 0 {
		out += "Draw4"
	} else if card.Flag&ColorChange != 0 {
		out += "ColorChange"
	}

	return fmt.Sprintf("%s", out)
}
