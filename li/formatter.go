package li

type FormatterConfig struct {
	DelaySeconds int
}

func (_ Provide) FormatterConfig(
	get GetConfig,
) FormatterConfig {
	var config struct {
		Formatter FormatterConfig
	}
	ce(get(&config))
	return config.Formatter
}
