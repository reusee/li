package li

import (
	"sync"

	"github.com/gdamore/tcell"
)

type Style = tcell.Style

type StyleConfig map[string]StyleSpec

type StyleSpec struct {
	FG        *int32
	BG        *int32
	Bold      *bool
	Underline *bool
}

func (s StyleSpec) ToFunc() StyleFunc {
	fn := SameStyle
	if s.FG != nil {
		fn = fn.SetFG(tcell.NewHexColor(*s.FG))
	}
	if s.BG != nil {
		fn = fn.SetBG(tcell.NewHexColor(*s.BG))
	}
	if s.Bold != nil {
		fn = fn.SetBold(*s.Bold)
	}
	if s.Underline != nil {
		fn = fn.SetUnderline(*s.Underline)
	}
	return fn
}

func (_ Provide) DefaultStyle(
	config StyleConfig,
) Style {
	style := tcell.StyleDefault
	if spec, ok := config["Default"]; ok {
		style = spec.ToFunc()(style)
	}
	return style
}

func (_ Provide) StyleConfig(
	getConfig GetConfig,
) StyleConfig {
	var config struct {
		Style StyleConfig
	}
	ce(getConfig(&config))
	return config.Style
}

type GetStyle func(keys ...string) StyleFunc

func (_ Provide) GetStyle(
	config StyleConfig,
) GetStyle {
	var specs sync.Map
	return func(keys ...string) StyleFunc {
		keys = append(keys, "Default")
		for _, key := range keys {
			v, ok := specs.Load(key)
			if ok {
				if fn, ok := v.(StyleFunc); ok {
					return fn
				} else if v == nil {
					continue
				}
			}
			spec, ok := config[key]
			if !ok {
				specs.Store(key, nil)
			} else {
				fn := spec.ToFunc()
				specs.Store(key, fn)
				return fn
			}
		}
		return SameStyle
	}
}
