package li

type ReadMode struct{}

var _ KeyStrokeHandler = new(ReadMode)

func (_ ReadMode) StrokeSpecs() any {
	return func(
		getConfig GetConfig,
	) []StrokeSpec {

		var config struct {
			ReadMode struct {
				SequenceCommand map[string]string
			}
		}
		ce(getConfig(&config))

		return strokeSpecsFromSequenceCommand(config.ReadMode.SequenceCommand)
	}
}
