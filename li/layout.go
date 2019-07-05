package li

import (
	"math"
	"reflect"
)

type Layout func(
	box Box,
) (
	max int,
	split func(n int) []ZBox,
)

var NamedLayouts = func() map[string]Layout {
	m := make(map[string]Layout)
	var f Layout
	v := reflect.ValueOf(f)
	t := reflect.TypeOf(f)
	for i := 0; i < t.NumMethod(); i++ {
		name := t.Method(i).Name
		var fn Layout
		fnValue := reflect.ValueOf(&fn)
		fnValue.Elem().Set(v.Method(i))
		m[name] = fn
	}
	return m
}()

func (_ Layout) VerticalSplit(box Box) (int, func(int) []ZBox) {
	return math.MaxInt32, func(n int) (ret []ZBox) {
		widths := split(box.Width(), n)
		left := box.Left
		for i := 0; i < n; i++ {
			ret = append(ret, ZBox{
				Box: Box{
					Left:   left,
					Right:  left + widths[i],
					Top:    box.Top,
					Bottom: box.Bottom,
				},
				Z: 0,
			})
			left += widths[i]
		}
		return
	}
}

func (_ Layout) HorizontalSplit(box Box) (int, func(int) []ZBox) {
	return math.MaxInt32, func(n int) (ret []ZBox) {
		heights := split(box.Height(), n)
		top := box.Top
		for i := 0; i < n; i++ {
			ret = append(ret, ZBox{
				Box: Box{
					Left:   box.Left,
					Right:  box.Right,
					Top:    top,
					Bottom: top + heights[i],
				},
				Z: 0,
			})
			top += heights[i]
		}
		return
	}
}

func (l Layout) BinarySplit(box Box) (int, func(int) []ZBox) {
	return math.MaxInt32, func(n int) (ret []ZBox) {
		if n <= 1 {
			return []ZBox{
				{
					Box: box,
					Z:   0,
				},
			}
		}
		n1 := n / 2
		n2 := n - n1
		var box1, box2 Box
		if box.Width()*10 > box.Height()*18 {
			box1 = Box{
				Left:   box.Left,
				Right:  box.Left + box.Width()/2,
				Top:    box.Top,
				Bottom: box.Bottom,
			}
			box2 = Box{
				Left:   box.Left + box.Width()/2,
				Right:  box.Right,
				Top:    box.Top,
				Bottom: box.Bottom,
			}
		} else {
			box1 = Box{
				Left:   box.Left,
				Right:  box.Right,
				Top:    box.Top,
				Bottom: box.Top + box.Height()/2,
			}
			box2 = Box{
				Left:   box.Left,
				Right:  box.Right,
				Top:    box.Top + box.Height()/2,
				Bottom: box.Bottom,
			}
		}
		_, split1 := l.BinarySplit(box1)
		_, split2 := l.BinarySplit(box2)
		return append(split1(n1), split2(n2)...)
	}
}

func (_ Layout) Stacked(box Box) (int, func(int) []ZBox) {
	return 1, func(n int) (ret []ZBox) {
		for i := 0; i < n; i++ {
			ret = append(ret, ZBox{
				Box: box,
				Z:   n - i,
			})
		}
		return
	}
}
