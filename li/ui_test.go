package li

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/gdamore/tcell"
)

func TestUIDescription(t *testing.T) {
	screenWidth := 80
	screenHeight := 25
	leftPanelWidth := 10

	tick := 0

	root := Rect(
		BGColor(0x202020),
		FGColor(0xC0C0C0),
		Fill(true),

		// left panel
		Rect(
			func(parent Box) Box {
				return Box{parent.Top, parent.Left, parent.Bottom, parent.Left + leftPanelWidth}
			},
			func(parent BGColor) BGColor {
				return BGColor(darkerOrLighterColor(Color(parent), 10))
			},
		),

		// top panel
		Rect(
			func(parent Box) Box {
				return Box{parent.Top, parent.Left + leftPanelWidth, parent.Top + 1, parent.Right}
			},
			func(parent BGColor) BGColor {
				return BGColor(darkerOrLighterColor(Color(parent), 20))
			},
			Bold(true),
			Underline(true),
		),

		// main area
		Rect(
			func(parent Box) Box {
				return Box{parent.Top + 1, parent.Left + leftPanelWidth, parent.Bottom, parent.Right}
			},
			func(parent BGColor) BGColor {
				return BGColor(darkerOrLighterColor(Color(parent), 30))
			},
			Margin(5),
			Padding(9, 11),

			// content
			func() Element {
				switch tick {

				case 0:
					return Text(
						"Hello, world!",
						"你好，世界！",
						"こんにちわ、世界！",
						FGColor(0xFF00CC),
					)

				case 1:
					return Text(
						"你好，世界！",
						"こんにちわ、世界！",
						"Hello, world!",
						FGColor(0xCC00FF),
					)

				}
				return Text("foo")
			},
		),
	)

	screen := tcell.NewSimulationScreen("")
	screen.SetSize(screenWidth, screenHeight)
	scope := NewScope(
		// defaults
		func() Box {
			return Box{0, 0, screenWidth, screenHeight}
		},
		func() Style {
			return tcell.StyleDefault
		},
		func() SetContent {
			return screen.SetContent
		},
	)
	renderAll(scope, root)

	tick++
	renderAll(scope, root)

}

func BenchmarkUI(b *testing.B) {
	tick := 0
	root := Rect(
		func(
			box Box,
		) (
			ret []Element,
		) {

			if tick%2 == 0 {
				for i := 0; i < box.Height(); i++ {
					ret = append(ret, Text(
						strings.Repeat("a", rand.Intn(box.Width())),
						Box{box.Left, box.Top + i, box.Top + i + 1, box.Right},
					))
				}

			} else {
				for i := 0; i < box.Height(); i++ {
					ret = append(ret, Text(
						strings.Repeat("b", rand.Intn(box.Width())),
						Box{box.Left, box.Top + i, box.Top + i + 1, box.Right},
					))
				}
			}

			return
		},
	)

	screenWidth := 80
	screenHeight := 25
	screen := tcell.NewSimulationScreen("")
	screen.SetSize(screenWidth, screenHeight)
	scope := NewScope(
		// defaults
		func() Box {
			return Box{0, 0, screenWidth, screenHeight}
		},
		func() Style {
			return tcell.StyleDefault
		},
		func() SetContent {
			return screen.SetContent
		},
	)
	renderAll(scope, root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tick++
		renderAll(scope, root)
	}

}
