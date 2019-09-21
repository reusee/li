package li

import (
	"github.com/gdamore/tcell"
)

type (
	Color = tcell.Color
)

var (
	HexColor = tcell.NewHexColor
	RGBColor = tcell.NewRGBColor
)

func towards128(x int32, n int32) (int32, int32) {
	if x > 128 {
		x -= n
		return x, -n
	}
	x += n
	return x, n
}

func darkerOrLighterStyle(style Style, n int32) Style {
	fg, bg, _ := style.Decompose()
	r, g, b := fg.RGB()
	mono := r == g && g == b
	r2, g2, b2 := bg.RGB()
	r2, d := towards128(r2, n)
	if mono {
		r += d
	}
	g2, d = towards128(g2, n)
	if mono {
		g += d
	}
	b2, d = towards128(b2, n)
	if mono {
		b += d
	}
	fg = tcell.NewRGBColor(r, g, b)
	bg = tcell.NewRGBColor(r2, g2, b2)
	return style.Foreground(fg).Background(bg)
}
