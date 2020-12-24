package li

import "sync/atomic"

type Overlay struct {
	ID               ID
	Element          Element
	KeyStrokeHandler KeyStrokeHandler
}

func (_ Provide) DefaultOverlays() []Overlay {
	return nil
}

var overlayID int64

type OverlayObject any

type PushOverlay func(
	obj OverlayObject,
) (
	id ID,
)

func (_ Provide) PushOverlay(
	run RunInMainLoop,
) PushOverlay {
	return func(
		obj OverlayObject,
	) (
		id ID,
	) {

		id = ID(atomic.AddInt64(&overlayID, 1))
		overlay := Overlay{
			ID: id,
		}
		if elem, ok := obj.(Element); ok {
			overlay.Element = elem
		}
		if handler, ok := obj.(KeyStrokeHandler); ok {
			overlay.KeyStrokeHandler = handler
		}
		run(func(
			scope Scope,
			overlays []Overlay,
			derive Derive,
		) {
			overlays = append(overlays, overlay)
			derive(
				func() []Overlay {
					return overlays
				},
			)
		})
		return
	}

}

type CloseOverlay func(
	id ID,
)

func (_ Provide) CloseOverlay(
	run RunInMainLoop,
) CloseOverlay {
	return func(
		id ID,
	) {
		run(func(
			scope Scope,
			overlays []Overlay,
			derive Derive,
		) {
			for i := 0; i < len(overlays); i++ {
				if overlays[i].ID == id {
					overlays = append(overlays[:i], overlays[i+1:]...)
				}
			}
			derive(
				func() []Overlay {
					return overlays
				},
			)
		})
	}
}
