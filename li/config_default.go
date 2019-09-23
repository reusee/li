package li

// create ~/.config/li-editor/config.toml to overwrite defaults

const DefaultConfig = `

[Scroll]
PaddingTop = 20
PaddingBottom = 20

[Mouse]
ScrollLines = 5

[Buffer]
ExpandTabs = true
TabWidth = 4

[UI]
StatusWidth = 20
JournalHeight = 2
MaxOutlineDistance = 2000

[ViewGroup]
Layouts = [
  'VerticalSplit',
  'HorizontalSplit',
  'BinarySplit',
  'Stacked',
]

  [[ViewGroup.Groups]]
  Layouts = [
    'Stacked',
    'BinarySplit',
    'VerticalSplit',
    'HorizontalSplit',
  ]

  [[ViewGroup.Groups]]
  Layouts = [
    'Stacked',
    'BinarySplit',
    'VerticalSplit',
    'HorizontalSplit',
  ]

[Style]

  [Style.Default]
  FG = 0xBBBBBB
  BG = 0x222222
  Bold = false
  Underline = false

  [Style.Highlight]
  FG = 0xAAFF00
  BG = 0x222222

[ReadMode]

  [ReadMode.SequenceCommand]

  'F2' = 'ToggleMacroRecording'

  'Rune[` + "`" + `]' = 'ToggleJournalHeight'
  'Rune[#]' = 'LineBegin'
  'Rune[$]' = 'LineEnd'

  'Rune[w]' = 'FocusPrevViewInGroup'
  'Rune[e]' = 'FocusNextViewInGroup'
  'Rune[U]' = 'PageUp'
  'Rune[u]' = 'UndoDuration1'
  'Rune[i]' = 'EnableEditMode'
  'Rune[O]' = 'EditNewLineAbove'
  'Rune[o]' = 'EditNewLineBelow'
  'Rune[{]' = 'PrevEmptyLine'
  'Rune[[]' = 'PrevDedentLine'
  'Rune[}]' = 'NextEmptyLine'
  'Rune[]]' = 'NextDedentLine'

  'Rune[a]' = 'Append'
  'Rune[d]' = 'Delete'
  'Rune[F]' = 'PrevRune'
  'Rune[f]' = 'NextRune'
  'Rune[G]' = 'ScrollAbsOrEnd'
  'Rune[g] Rune[g]' = 'ScrollAbsOrHome'
  'Rune[h]' = 'MoveLeft'
  'Rune[j]' = 'MoveDown'
  'Rune[k]' = 'MoveUp'
  'Rune[l]' = 'MoveRight'
  'Rune[;]' = 'Imitate'

  'Rune[z] Rune[t]' = 'ScrollCursorToUpper' 
  'Rune[z] Rune[z]' = 'ScrollCursorToMiddle' 
  'Rune[z] Rune[b]' = 'ScrollCursorToLower' 
  'Rune[x]' = 'DeleteRune' 
  'Rune[c]' = 'Change' 
  'Rune[c] Rune[w]' = 'ChangeToWordEnd' 
  'Rune[v]' = 'ToggleSelection' 
  'Rune[b]' = 'ShowViewSwitcher' 
  'Rune[M]' = 'PageDown'
  'Rune[/]' = 'ShowSearchDialog'

  'Rune[,] Rune[q]' = 'CloseView'
  'Rune[,] Rune[w]' = 'SyncViewToFile'
  'Rune[,] Rune[t]' = 'ChoosePathAndLoad'
  'Rune[,] Rune[f]' = 'NextLineWithRune'
  'Rune[,] Rune[g]' = 'NextViewGroupLayout'
  'Rune[,] Rune[v]' = 'NextViewLayout'

  'Rune[,] Rune[N]' = 'CurrentTime'

  'Rune[.] Rune[g]' = 'PrevViewGroupLayout'
  'Rune[.] Rune[f]' = 'PrevLineWithRune'
  'Rune[.] Rune[v]' = 'PrevViewLayout'

  'Alt+Rune[u]' = 'RedoLatest'

  'Ctrl+U' = 'Undo'
  'Ctrl+O' = 'ShowCommandPalette'

[EditMode]
DisableSequence = "kd"

  [EditMode.SequenceCommand]

  'Esc' = 'DisableEditMode'

  'Ctrl+O' = 'ShowCommandPalette'

[Undo]
DurationMS1 = 3000

[Debug]
Verbose = false

[LanguageServerProtocol]
Enable = false

`
