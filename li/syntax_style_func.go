package li

type SyntaxStyles struct {
	Keyword StyleFunc
	Type    StyleFunc
	Literal StyleFunc
	Builtin StyleFunc
	Comment StyleFunc
}

func (_ Provide) DefaultSyntaxStyles(
	colors Colors,
) SyntaxStyles {
	return SyntaxStyles{
		Keyword: SetFG(colors.Red),
		Type:    SetFG(colors.Blue),
		Literal: SetFG(colors.Green),
		Builtin: SetFG(colors.Yellow),
		Comment: SetFG(colors.Yellow),
	}
}
