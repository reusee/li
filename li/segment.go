package li

type Segment struct {
	lines []*Line
}

type Segments []*Segment

func (s Segments) Len() (l int) {
	for _, segment := range s {
		l += len(segment.lines)
	}
	return
}

func (s Segments) Sub(start int, end int) Segments {
	if start < 0 {
		start = 0
	}
	if l := s.Len(); end < 0 || end > l {
		end = l
	}

	n := end
	i := 0
	var ret Segments
	for n > 0 && i < len(s) {
		segment := s[i]
		if n < len(segment.lines) {
			// split
			segment = &Segment{
				lines: segment.lines[:n],
			}
			ret = append(ret, segment)
			break
		} else {
			ret = append(ret, segment)
			i++
			n -= len(segment.lines)
		}
	}

	n = start
	i = 0
	for n > 0 && i < len(ret) {
		segment := ret[i]
		if n >= len(segment.lines) {
			ret = ret[1:]
			n -= len(segment.lines)
		} else {
			// split
			newSegment := &Segment{
				lines: segment.lines[n:],
			}
			ret[0] = newSegment
			break
		}
	}

	return ret
}
