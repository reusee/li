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

	black = HexColor(0)
)

func towards128(x int32, n int32) (int32, int32) {
	if x > 128 {
		x -= n
		return x, -n
	}
	x += n
	return x, n
}

func darkerOrLighterColor(color Color, n int32) Color {
	r, g, b := color.RGB()
	r, _ = towards128(r, n)
	g, _ = towards128(g, n)
	b, _ = towards128(b, n)
	return tcell.NewRGBColor(r, g, b)
}

func darkerOrLighterColor2(fg, bg Color, n int32) (Color, Color) {
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
	return tcell.NewRGBColor(r, g, b), tcell.NewRGBColor(r2, g2, b2)
}

func darkerOrLighterStyle(style Style, n int32) Style {
	fg, bg, _ := style.Decompose()
	fg, bg = darkerOrLighterColor2(fg, bg, n)
	return style.Foreground(fg).Background(bg)
}
