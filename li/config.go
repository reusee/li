package li

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/reusee/e4"
	"github.com/reusee/toml"
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

	userConfig, err := ioutil.ReadFile(filepath.Join(configDir, "config.toml"))
	if err != nil && !os.IsNotExist(err) {
		ce(err, e4.NewInfo("open config.toml"))
	}

	defaultConfig := []byte(DefaultConfig)

	get = func(target any) error {
		if err := toml.Unmarshal(defaultConfig, target); err != nil {
			return err
		}
		if err := toml.Unmarshal(userConfig, target); err != nil {
			return err
		}
		return nil
	}

	return
}
