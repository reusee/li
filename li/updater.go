package li

type Updater struct {
	Args map[string][]any
}

func (u *Updater) Update(key string, fn func(), args ...any) {
	changed := false
	lastArgs, ok := u.Args[key]
	if !ok {
		changed = true
	} else if len(args) != len(lastArgs) {
		changed = true
	} else {
		for i := 0; i < len(args); i++ {
			if args[i] != lastArgs[i] {
				changed = true
				break
			}
		}
	}
	if !changed {
		return
	}
	fn()
	u.Args[key] = args
}

func (u *Updater) ResetAll() {
	for k := range u.Args {
		delete(u.Args, k)
	}
}

func NewUpdater() *Updater {
	return &Updater{
		Args: make(map[string][]any),
	}
}

func (_ Provide) GlobalUpdater() *Updater {
	return NewUpdater()
}
