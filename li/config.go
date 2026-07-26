package li

import (
	"encoding/json"

	"os"
	"path/filepath"

	"github.com/reusee/e4"
)

type (
	GetConfig func(target any) error
	ConfigDir string
)

func getConfigDir() string {
	configDir, err := os.UserConfigDir()
	ce(err)
	configDir = filepath.Join(configDir, "li-editor")
	_, err = os.Stat(configDir)
	if os.IsNotExist(err) {
		ce(os.Mkdir(configDir, 0755))
	}
	return configDir
}

func (_ Provide) Config() (
	get GetConfig,
	dir ConfigDir,
) {

	configDir := getConfigDir()
	dir = ConfigDir(configDir)

	userConfig, err := os.ReadFile(filepath.Join(configDir, "config.json"))
	if err != nil && !os.IsNotExist(err) {
		ce(err, e4.NewInfo("open config.json"))
	}

	defaultConfig := []byte(DefaultConfig)

	get = func(target any) error {
		if err := json.Unmarshal(defaultConfig, target); err != nil {
			return err
		}
		if err := json.Unmarshal(userConfig, target); err != nil {
			return err
		}
		return nil
	}

	return
}
