package li

type GoLexicalStainer struct {
}

func (s *GoLexicalStainer) Line() any {
	return func(
		moment *Moment,
		line LineNumber,
		getStyle GetStyle,
	) []*Color {
		//TODO
		return nil
	}
}
