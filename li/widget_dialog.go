package li

type WidgetDialog struct {
	OnKey   any
	Element Element
}

var _ Element = WidgetDialog{}

var _ KeyStrokeHandler = WidgetDialog{}

func (w WidgetDialog) RenderFunc() any {
	return func(
		scope Scope,
	) {
		renderAll(scope, w.Element)
	}
}

func (w WidgetDialog) StrokeSpecs() any {
	return func() []StrokeSpec {
		return []StrokeSpec{
			{
				Predict: func() bool {
					return w.OnKey != nil
				},
				Func: func(ev KeyEvent, scope Scope) {
					scope.Call(w.OnKey)
				},
			},
		}
	}
}
