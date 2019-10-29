package li

import "time"

type (
	RunInMainLoop    func(fn any)
	ContinueMainLoop func()

	RenderTimer struct {
		*time.Timer
	}
	ResetRenderTimer func()
)

var renderTimeout = time.Millisecond * 10

func (_ Provide) Loop() (
	cont ContinueMainLoop,
	renderTimer RenderTimer,
	resetRenderTimer ResetRenderTimer,
) {

	timer := time.NewTimer(renderTimeout)
	renderTimer = RenderTimer{
		Timer: timer,
	}
	resetRenderTimer = func() {
		timer.Reset(renderTimeout)
	}

	cont = func() {
		resetRenderTimer()
	}

	return
}

type evLoopBegin struct{}

var EvLoopBegin = new(evLoopBegin)

type evLoopEnd struct{}

var EvLoopEnd = new(evLoopEnd)
