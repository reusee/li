package li

import (
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/junegunn/fzf/src/util"
)

func ShowFileChooser(scope Scope, cb func(string)) {

	// states
	type Candidate struct {
		Path     string
		MatchLen int
		Score    int
	}

	type Result struct {
		candidates []Candidate
		maxLength  int
	}
	cache := make(map[string]Result)

	updateCandidates := func(
		scope Scope,
		runes []rune,
	) (
		candidates []Candidate,
		maxLength int,
	) {

		if res, ok := cache[string(runes)]; ok {
			candidates = res.candidates
			maxLength = res.maxLength
			return
		}
		defer func() {
			cache[string(runes)] = Result{
				candidates: candidates,
				maxLength:  maxLength,
			}
		}()

		// reset state
		candidates = candidates[:0]
		maxLength = 0

		// match
		var path string
		if len(runes) >= 1 && runes[0] == '/' {
			// absolute path
			path = string(runes)
		} else if len(runes) >= 2 && runes[0] == '~' && runes[1] == '/' {
			// home dir
			homeDir, err := os.UserHomeDir()
			ce(err)
			path = homeDir + string(runes[1:])
		} else {
			// relative path
			var curView CurrentView
			scope.Assign(&curView)
			cur := curView()
			if cur != nil {
				path = filepath.Join(
					filepath.Dir(cur.Buffer.Path),
					string(runes),
				)
			} else {
				path = string(runes)
			}
		}
		path, err := filepath.Abs(path)
		if err != nil {
			return
		}
		if path != "/" && len(runes) > 0 && runes[len(runes)-1] == '/' {
			path += "/"
		} else if path != "/" && len(runes) == 0 {
			path += "/"
		}
		patterns := splitDir(path)

		var scan func(
			path string,
			patterns []string,
			matchLen int,
			score int,
		)
		scan = func(
			path string,
			patterns []string,
			matchLen int,
			score int,
		) {
			if len(patterns) == 0 {
				return
			}

			patternRunes := []rune(patterns[0])
			f, err := os.Open(path)
			if err != nil {
				return
			}
			defer f.Close()

			type Scan struct {
				fn    func()
				score int
			}
			var scans []Scan

			for {
				infos, err := f.Readdir(256)
				for _, info := range infos {
					info := info
					name := info.Name()
					fullPath := filepath.Join(path, name)
					if w := displayWidth(fullPath); w > maxLength {
						maxLength = w
					}
					chars := util.RunesToChars([]rune(name))
					matched, l, s := fuzzyMatched(patternRunes, &chars)
					if !matched {
						continue
					}
					if len(patterns) == 1 {
						// final
						candidates = append(candidates, Candidate{
							Path:     fullPath,
							MatchLen: matchLen + l,
							Score:    score + s,
						})

					} else {
						// descend
						if info.IsDir() {
							scans = append(scans, Scan{
								fn: func() {
									scan(
										fullPath,
										patterns[1:],
										matchLen+displayWidth(name)+1,
										score+s,
									)
								},
								score: score + s,
							})
						}
					}

				}

				if is(err, io.EOF) {
					break
				} else if err != nil {
					return
				}
			}

			sort.SliceStable(scans, func(i, j int) bool {
				return scans[i].score > scans[j].score
			})
			if len(scans) > 30 {
				scans = scans[:30]
			}
			for _, scan := range scans {
				scan.fn()
			}

		}
		scan("/", patterns, 1, 0)

		sort.SliceStable(candidates, func(i, j int) bool {
			c1 := candidates[i]
			c2 := candidates[j]
			if c1.Score != c2.Score {
				return c1.Score > c2.Score
			}
			return c1.MatchLen < c2.MatchLen
		})

		return
	}
	candidates, maxLength := updateCandidates(scope, nil)

	var id ID
	dialog := &SelectionDialog{

		Title: "Choose File",

		OnClose: func(_ Scope) {
			scope.Sub(&id).Call(CloseOverlay)
		},

		OnSelect: func(_ Scope, id ID) {
			scope.Sub(&id).Call(CloseOverlay)
			if int(id) < len(candidates) {
				cb(candidates[id].Path)
			}
		},

		OnUpdate: func(scope Scope, runes []rune) (ids []ID, maxLen int, initIndex int) {
			candidates, maxLength = updateCandidates(scope, runes)
			maxLen = maxLength
			for i := range candidates {
				ids = append(ids, ID(i))
			}
			return
		},

		CandidateElement: func(scope Scope, id ID) Element {
			var box Box
			var focus ID
			var style Style
			var getStyle GetStyle
			scope.Assign(&box, &focus, &style, &getStyle)
			s := style
			if id == focus {
				hlStyle := getStyle("Highlight")(s)
				fg, _, _ := hlStyle.Decompose()
				s = s.Foreground(fg)
			}
			candidate := candidates[id]
			return Text(
				box,
				candidate.Path,
				s,
				OffsetStyleFunc(func(i int) StyleFunc {
					fn := SameStyle
					if i < candidate.MatchLen {
						fn = fn.SetUnderline(true)
					} else {
						fn = fn.SetUnderline(false)
					}
					return fn
				}),
			)
		},
	}

	overlay := OverlayObject(dialog)
	scope.Sub(&overlay).Call(PushOverlay, &id)
}
