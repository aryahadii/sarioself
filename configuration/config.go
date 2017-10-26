package configuration

import (
	"github.com/spf13/viper"
)

var (
	// SarioselfConfigPath is path of config file
	SarioselfConfigPath = "config.yaml"

	// SarioselfConfig is config of project
	SarioselfConfig *viper.Viper
)

// LoadConfig loads Sarioself's config file from SarioselfConfigPath
func LoadConfig() error {
	SarioselfConfig = viper.New()
	SarioselfConfig.SetConfigFile(SarioselfConfigPath)
	if err := SarioselfConfig.ReadInConfig(); err != nil {
		return err
	}

	SarioselfConfig.SetDefault("address", "localhost:8000")
	SarioselfConfig.SetDefault("debug", true)

	return nil
}
