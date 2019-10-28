package li

import "time"

type (
	RunInMainLoop func(fn any)
	RenderTimer   struct {
		*time.Timer
	}
	ResetRenderTimer func()
)

type LoopStep func()

var renderTimeout = time.Millisecond * 10

func (_ Provide) Loop() (
	step LoopStep,
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

	step = func() {
		resetRenderTimer()
	}

	return
}

type evLoopBegin struct{}

var EvLoopBegin = new(evLoopBegin)

type evLoopEnd struct{}

var EvLoopEnd = new(evLoopEnd)
