package li

type StyleFunc func(style Style) Style

var SameStyle = StyleFunc(func(style Style) Style {
	return style
})

func SetFG(color Color) StyleFunc {
	return func(style Style) Style {
		return style.Foreground(color)
	}
}

func SetBG(color Color) StyleFunc {
	return func(style Style) Style {
		return style.Background(color)
	}
}

func SetBold(bold bool) StyleFunc {
	return func(style Style) Style {
		return style.Bold(bold)
	}
}

func SetUnderline(underline bool) StyleFunc {
	return func(style Style) Style {
		return style.Underline(underline)
	}
}

func (s StyleFunc) SetFG(color Color) StyleFunc {
	return func(style Style) Style {
		style = s(style)
		return style.Foreground(color)
	}
}

func (s StyleFunc) SetBG(color Color) StyleFunc {
	return func(style Style) Style {
		style = s(style)
		return style.Background(color)
	}
}

func (s StyleFunc) SetBold(bold bool) StyleFunc {
	return func(style Style) Style {
		style = s(style)
		return style.Bold(bold)
	}
}

func (s StyleFunc) SetUnderline(underline bool) StyleFunc {
	return func(style Style) Style {
		style = s(style)
		return style.Underline(underline)
	}
}

func (s StyleFunc) And(f2 StyleFunc) StyleFunc {
	return func(style Style) Style {
		return f2(s(style))
	}
}

//TODO set darker or lighter
