package li

import (
	"reflect"
	"sort"
	"sync"
)

type (
	Link          func(left, right any)
	LinkedOne     func(o, target any)
	LinkedAll     func(o, target any)
	DropLink      func(left, right any)
	DropLinked    func(o any)
	DropLinkedOne func(o any)
)

func (_ Provide) LinkFuncs() (
	link Link,
	linkedOne LinkedOne,
	linkedAll LinkedAll,
	dropLink DropLink,
	drop DropLinked,
	dropOne DropLinkedOne,
) {

	var serial int

	links := make(map[any]map[reflect.Type]map[any]int)
	var l sync.RWMutex

	save := func(left, right any, n int) {
		m, ok := links[left]
		if !ok {
			m = make(map[reflect.Type]map[any]int)
			links[left] = m
		}
		t := reflect.TypeOf(right)
		m2, ok := m[t]
		if !ok {
			m2 = make(map[any]int)
			m[t] = m2
		}
		links[left][t][right] = n
	}

	del := func(left, right any) {
		m, ok := links[left]
		if !ok {
			return
		}
		t := reflect.TypeOf(right)
		m2, ok := m[t]
		if !ok {
			return
		}
		delete(m2, right)
	}

	link = func(left, right any) {
		l.Lock()
		defer l.Unlock()
		n := serial
		serial++
		save(left, right, n)
		save(right, left, n)
	}

	linkedOne = func(o, target any) {
		l.RLock()
		defer l.RUnlock()
		t := reflect.TypeOf(target).Elem()
		m, ok := links[o]
		if !ok {
			return
		}
		m2, ok := m[t]
		if !ok {
			return
		}
		if len(m2) == 0 {
			return
		}
		var rights []any
		for right := range m2 {
			rights = append(rights, right)
		}
		sort.SliceStable(rights, func(i, j int) bool {
			return m2[rights[i]] > m2[rights[j]]
		})
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(rights[0]))
	}

	linkedAll = func(o, target any) {
		l.RLock()
		defer l.RUnlock()
		t := reflect.TypeOf(target).Elem().Elem()
		m, ok := links[o]
		if !ok {
			return
		}
		m2, ok := m[t]
		if !ok {
			return
		}
		if len(m2) == 0 {
			return
		}
		var rights []any
		for right := range m2 {
			rights = append(rights, right)
		}
		sort.SliceStable(rights, func(i, j int) bool {
			return m2[rights[i]] > m2[rights[j]]
		})
		slice := reflect.New(reflect.TypeOf(target).Elem()).Elem()
		for _, right := range rights {
			slice = reflect.Append(slice, reflect.ValueOf(right))
		}
		reflect.ValueOf(target).Elem().Set(slice)
	}

	drop = func(left any) {
		l.Lock()
		defer l.Unlock()
		for _, m := range links[left] {
			for right := range m {
				del(right, left)
			}
		}
		delete(links, left)
	}

	dropLink = func(left any, right any) {
		l.Lock()
		defer l.Unlock()
		m, ok := links[left]
		if ok {
			m2, ok := m[reflect.TypeOf(right)]
			if ok {
				delete(m2, right)
			}
		}
		m, ok = links[right]
		if ok {
			m2, ok := m[reflect.TypeOf(left)]
			if ok {
				delete(m2, left)
			}
		}
	}

	type Info struct {
		Obj any
		N   int
	}

	dropOne = func(left any) {
		l.Lock()
		defer l.Unlock()
		m, ok := links[left]
		if !ok {
			return
		}
		var infos []Info
		for _, m2 := range m {
			for right, n := range m2 {
				infos = append(infos, Info{
					Obj: right,
					N:   n,
				})
			}
		}
		if len(infos) == 0 {
			return
		}
		sort.SliceStable(infos, func(i, j int) bool {
			return infos[i].N > infos[j].N
		})
		del(left, infos[0].Obj)
		del(infos[0].Obj, left)
	}

	return
}
