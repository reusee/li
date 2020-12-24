package li

import "github.com/gdamore/tcell"

type Colors struct {
	Black   tcell.Color
	Red     tcell.Color
	Green   tcell.Color
	Yellow  tcell.Color
	Blue    tcell.Color
	Magenta tcell.Color
	Cyan    tcell.Color
	White   tcell.Color

	LightBlack   tcell.Color
	LightRed     tcell.Color
	LightGreen   tcell.Color
	LightYellow  tcell.Color
	LightBlue    tcell.Color
	LightMagenta tcell.Color
	LightCyan    tcell.Color
	LightWhite   tcell.Color
}

func (_ Provide) DefaultColors() Colors {
	return Colors{
		Black:   HexColor(0x555555),
		Red:     HexColor(0xff8272),
		Green:   HexColor(0xb4fa72),
		Yellow:  HexColor(0xfefdc2),
		Blue:    HexColor(0xadd5fe),
		Magenta: HexColor(0xff8ffd),
		Cyan:    HexColor(0xd0d1fe),
		White:   HexColor(0xf3f3f3),

		LightBlack:   HexColor(0x666666),
		LightRed:     HexColor(0xffc4bd),
		LightGreen:   HexColor(0xd6fcb9),
		LightYellow:  HexColor(0xfefdd5),
		LightBlue:    HexColor(0xc1e3fe),
		LightMagenta: HexColor(0xffb1fe),
		LightCyan:    HexColor(0xe5e6fe),
		LightWhite:   HexColor(0xfeffff),
	}
}
