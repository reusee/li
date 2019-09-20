package li

import "time"

func (_ Command) CurrentTime() (spec CommandSpec) {
	spec.Desc = "show current time"
	spec.Func = func(
		appendJournal AppendJournal,
	) {
		appendJournal("%s", time.Now().Format("2006-01-02 15:04:05"))
	}
	return
}
