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

func PushOverlay(
	obj OverlayObject,
	onNext OnNext,
) (id ID) {
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
	onNext(EvLoopBegin, func(
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

func CloseOverlay(
	id ID,
	onNext OnNext,
) {
	onNext(EvLoopBegin, func(
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
