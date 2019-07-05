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
	Bold      bool
	Underline bool
}

func (s StyleSpec) ToStyle() Style {
	style := tcell.StyleDefault
	if s.FG != nil {
		style = style.Foreground(tcell.NewHexColor(*s.FG))
	}
	if s.BG != nil {
		style = style.Background(tcell.NewHexColor(*s.BG))
	}
	style = style.Bold(s.Bold)
	style = style.Underline(s.Underline)
	return style
}

func (_ Provide) DefaultStyle(
	config StyleConfig,
) Style {
	if spec, ok := config["Default"]; ok {
		return spec.ToStyle()
	}
	return tcell.StyleDefault
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

type GetStyle func(keys ...string) Style

func (_ Provide) GetStyle(
	config StyleConfig,
) GetStyle {
	var specs sync.Map
	return func(keys ...string) Style {
		keys = append(keys, "Default")
		for _, key := range keys {
			v, ok := specs.Load(key)
			if ok {
				if style, ok := v.(Style); ok {
					return style
				} else if v == nil {
					continue
				}
			}
			spec, ok := config[key]
			if !ok {
				specs.Store(key, nil)
			} else {
				style := spec.ToStyle()
				specs.Store(key, style)
				return style
			}
		}
		return tcell.StyleDefault
	}
}
