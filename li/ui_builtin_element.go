package li

import (
	"fmt"
	"math"
)

// Rect

type _Rect struct {
	UIDesc
}

var _ Element = _Rect{}

func (r _Rect) RenderFunc() any {
	return func(
		scope Scope,
		parentBox Box,
		parentStyle Style,
		setContent SetContent,
	) {

		box := parentBox
		style := parentStyle
		var children []Element
		var marginLeft, marginRight, marginTop, marginBottom int
		var paddingLeft, paddingRight, paddingTop, paddingBottom int
		fill := false

		r.IterSpecs(scope, func(v interface{}) {
			switch v := v.(type) {

			case Box:
				box = v

			case FGColor:
				style = style.Foreground(Color(v))

			case BGColor:
				style = style.Background(Color(v))

			case Style:
				style = v

			case Element:
				if v != nil {
					children = append(children, v)
				}

			case []Element:
				for _, elem := range v {
					if elem != nil {
						children = append(children, elem)
					}
				}

			case Bold:
				style = style.Bold(bool(v))

			case Underline:
				style = style.Underline(bool(v))

			case Fill:
				fill = bool(v)

			case _Margin:
				if len(v) == 1 {
					marginTop = v[0]
					marginBottom = v[0]
					marginLeft = v[0]
					marginRight = v[0]
				} else if len(v) == 2 {
					marginTop = v[0]
					marginBottom = v[0]
					marginLeft = v[1]
					marginRight = v[1]
				} else if len(v) == 3 {
					marginTop = v[0]
					marginLeft = v[1]
					marginRight = v[1]
					marginBottom = v[2]
				} else if len(v) == 4 {
					marginTop = v[0]
					marginRight = v[1]
					marginBottom = v[2]
					marginLeft = v[3]
				} else {
					panic(me(nil, "bad margin: %q\n", v))
				}

			case _Padding:
				if len(v) == 1 {
					paddingTop = v[0]
					paddingBottom = v[0]
					paddingLeft = v[0]
					paddingRight = v[0]
				} else if len(v) == 2 {
					paddingTop = v[0]
					paddingBottom = v[0]
					paddingLeft = v[1]
					paddingRight = v[1]
				} else if len(v) == 3 {
					paddingTop = v[0]
					paddingLeft = v[1]
					paddingRight = v[1]
					paddingBottom = v[2]
				} else if len(v) == 4 {
					paddingTop = v[0]
					paddingRight = v[1]
					paddingBottom = v[2]
					paddingLeft = v[3]
				} else {
					panic(me(nil, "bad padding: %q\n", v))
				}

			default:
				panic(me(nil, "unknown spec %#v\n", v))
			}
		})

		marks := make([]bool, box.Width()*box.Height())
		set := func(x int, y int, mainc rune, combc []rune, style Style) {
			idx := (y-box.Top)*box.Width() + (x - box.Left)
			if idx >= 0 && idx < len(marks) {
				marks[idx] = true
			}
			setContent(x, y, mainc, combc, style)
		}

		fg, bg, _ := style.Decompose()
		childBox := Box{
			Top:    box.Top + marginTop + paddingTop,
			Left:   box.Left + marginLeft + paddingLeft,
			Right:  box.Right + marginRight + paddingRight,
			Bottom: box.Bottom + marginBottom + paddingBottom,
		}
		childScope := scope.Sub(
			func() (Box, SetContent) {
				return childBox, set
			},
			func() (Style, FGColor, BGColor) {
				return style, FGColor(fg), BGColor(bg)
			},
		)
		renderAll(childScope, children...)

		if fill {
			y := box.Top + marginTop
			maxY := box.Bottom - marginBottom
			for ; y < maxY; y++ {
				x := box.Left + marginLeft
				maxX := box.Right - marginRight
				for ; x < maxX; x++ {
					idx := (y-box.Top)*box.Width() + (x - box.Left)
					if !marks[idx] {
						setContent(x, y, ' ', nil, style)
					}
				}
			}
		}

	}
}

func Rect(specs ...any) _Rect {
	return _Rect{
		UIDesc: NewUIDesc(specs),
	}
}

// Text

type _Text struct {
	UIDesc
}

var _ Element = _Text{}

type OffsetStyleFunc func(int) Style

func (t _Text) RenderFunc() any {
	return func(
		parentBox Box,
		parentStyle Style,
		setContent SetContent,
		scope Scope,
	) {

		box := parentBox
		style := parentStyle
		var lines []string
		align := AlignLeft
		var paddingLeft, paddingRight, paddingTop, paddingBottom int
		var offsetStyleFunc OffsetStyleFunc
		fill := false

		t.IterSpecs(scope, func(v interface{}) {
			switch v := v.(type) {

			case Box:
				box = v

			case FGColor:
				style = style.Foreground(Color(v))

			case BGColor:
				style = style.Background(Color(v))

			case Style:
				style = v

			case Bold:
				style = style.Bold(bool(v))

			case Underline:
				style = style.Underline(bool(v))

			case Fill:
				fill = bool(v)

			case string:
				lines = append(lines, v)

			case []string:
				lines = append(lines, v...)

			case Align:
				align = v

			case _Padding:
				if len(v) == 1 {
					paddingTop = v[0]
					paddingBottom = v[0]
					paddingLeft = v[0]
					paddingRight = v[0]
				} else if len(v) == 2 {
					paddingTop = v[0]
					paddingBottom = v[0]
					paddingLeft = v[1]
					paddingRight = v[1]
				} else if len(v) == 3 {
					paddingTop = v[0]
					paddingLeft = v[1]
					paddingRight = v[1]
					paddingBottom = v[2]
				} else if len(v) == 4 {
					paddingTop = v[0]
					paddingRight = v[1]
					paddingBottom = v[2]
					paddingLeft = v[3]
				} else {
					panic(me(nil, "bad padding: %q\n", v))
				}

			case OffsetStyleFunc:
				offsetStyleFunc = v

			default:
				panic(me(nil, "unknown spec %#v\n", v))
			}
		})

		maxY := box.Bottom - paddingBottom
		for i, line := range lines {
			runes := []rune(line)
			var left int
			switch align {
			case AlignLeft:
				left = box.Left + paddingLeft
			case AlignRight:
				left = box.Right - paddingRight - runesDisplayWidth(runes)
			case AlignCenter:
				left = (box.Left+box.Right)/2 - runesDisplayWidth(runes)/2
			}
			for left < box.Left && len(runes) > 0 {
				r := runes[0]
				runes = runes[1:]
				left += runeWidth(r)
			}
			y := box.Top + paddingTop + i
			for runeIdx, r := range runes {
				if left >= box.Right {
					break
				}
				if y > maxY {
					continue
				}
				s := style
				if offsetStyleFunc != nil {
					s = offsetStyleFunc(runeIdx)
				}
				setContent(left, y, r, nil, s)
				left += runeWidth(r)
			}
			if fill {
				for left < box.Right {
					setContent(left, y, ' ', nil, style)
					left++
				}
			}
		}

	}
}

func Text(specs ...any) _Text {
	return _Text{
		UIDesc: NewUIDesc(specs),
	}
}

// FrameBuffer

type FrameBufferCell struct {
	Rune  rune
	Style Style
}

type FrameBuffer struct {
	Cells  []*FrameBufferCell
	Left   int
	Top    int
	Width  int
	Height int
}

var _ Element = new(FrameBuffer)

func NewFrameBuffer(box Box) *FrameBuffer {
	return &FrameBuffer{
		Cells:  make([]*FrameBufferCell, box.Width()*box.Height()),
		Left:   box.Left,
		Top:    box.Top,
		Width:  box.Width(),
		Height: box.Height(),
	}
}

func (f *FrameBuffer) RenderFunc() any {
	return func(
		box Box,
		set SetContent,
	) {
		targetWidth := box.Width()
		targetHeight := box.Height()
		for y := 0; y < f.Height; y++ {
			for x := 0; x < f.Width; x++ {
				cell := f.Cells[y*f.Width+x]
				if cell == nil {
					continue
				}
				if x >= targetWidth || y >= targetHeight {
					continue
				}
				set(box.Left+x, box.Top+y, cell.Rune, nil, cell.Style)
			}
		}
	}
}

func (f *FrameBuffer) SetContent(x int, y int, mainc rune, combc []rune, style Style) {
	x -= f.Left
	y -= f.Top
	i := y*f.Width + x
	if i >= len(f.Cells) {
		return
		//panic("bad pos")
	}
	f.Cells[i] = &FrameBufferCell{
		Rune:  mainc,
		Style: style,
	}
}

// VerticalScroll

type _VerticalScroll struct {
	element Element
	offset  int
}

func VerticalScroll(e Element, offset int) _VerticalScroll {
	return _VerticalScroll{
		element: e,
		offset:  offset,
	}
}

var _ Element = _VerticalScroll{}

func (v _VerticalScroll) RenderFunc() any {
	return func(
		box Box,
		setContent SetContent,
		scope Scope,
		style Style,
	) {
		elemBox := Box{
			Left:   box.Left,
			Right:  box.Right,
			Top:    box.Top,
			Bottom: math.MaxInt32,
		}
		maxY := box.Top
		type Cell struct {
			Rune  rune
			Style Style
		}
		cells := make(map[int]map[int]Cell)
		set := func(x int, y int, mainc rune, combc []rune, style Style) {
			if y > maxY {
				maxY = y
			}
			line, ok := cells[y]
			if !ok {
				line = make(map[int]Cell)
				cells[y] = line
			}
			line[x] = Cell{
				Rune:  mainc,
				Style: style,
			}
		}
		scope.Sub(func() (Box, SetContent) {
			return elemBox, set
		}).Call(v.element.RenderFunc())
		fromY := box.Top + v.offset - box.Height()/2
		if fromY < box.Top {
			fromY = box.Top
		}
		numTopCrop := fromY - box.Top
		for i := 0; i < box.Height(); i++ {
			y := fromY + i
			for x, cell := range cells[y] {
				setContent(x, y-numTopCrop, cell.Rune, nil, cell.Style)
			}
		}
		numBottomCrop := maxY - (fromY + box.Height()) + 1
		if numTopCrop > 0 {
			s := darkerOrLighterStyle(style, 15).Bold(true)
			runes := []rune(fmt.Sprintf(" %d.. ", numTopCrop))
			for i, r := range runes {
				setContent(box.Left+i, box.Top, r, nil, s)
			}
		}
		if numBottomCrop > 0 {
			s := darkerOrLighterStyle(style, 15).Bold(true)
			runes := []rune(fmt.Sprintf(" %d.. ", numBottomCrop))
			for i, r := range runes {
				setContent(box.Left+i, box.Bottom-1, r, nil, s)
			}
		}
	}
}
